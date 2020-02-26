package k8s

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "yager/handler"
	"yager/model"
	"yager/pkg/errno"
	"yager/pkg/logger"
	"yager/util"

	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podAnnoType  = "cloud.graviti.cn/type"
	podAnnoEnv   = "cloud.graviti.cn/env"
	podAnnoOrgID = "cloud.graviti.cn/orgid"
	podAnnoAppID = "cloud.graviti.cn/appid"
	//podAnnoIngressBandwidth = "kubernetes.io/ingress-bandwidth"
	//podAnnoEgressBandwidth  = "kubernetes.io/egress-bandwidth"
)

//GetDeployment get deployment instances from k8s
func GetDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	//
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	cs, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	begin := time.Now()
	dps, err := cs.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get deployment info.")
		return
	}
	logger.MetricsEmit(
		"k8s.get_dps",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == err,
	)

	SendResponse(ctx, errno.OK, dps)
}

// CreateDeployment create deployment instance
func CreateDeployment(ctx *gin.Context) {

	log := logger.RuntimeLog

	var ds *model.Deployment
	if err := ctx.BindJSON(&ds); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
	}

	appDs := ds.AppMeta
	fmt.Printf("name: %s %s %s \n", appDs.Name, appDs.ZoneName, appDs.AppID)
	appDs.Namespace = fmt.Sprintf("%s-app", strings.ToLower(appDs.Env))

	cs, err := GetClientByAzCode(appDs.ZoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	annotations := map[string]string{
		"cloud.graviti.cn/zone":  appDs.ZoneName,
		"cloud.graviti.cn/orgid": appDs.OrgID,
		"cloud.graviti.cn/appid": appDs.AppID,
		"cloud.graviti.cn/env":   appDs.Env,
	}

	for _, anno := range ds.DsSpec.Annotations {
		annotations[anno.Key] = anno.Value
	}

	//jvmRatio, err := group.JVMRatio()
	//if err != nil {
	//	log.WithError(err).Errorf("Failed get jvmRatio of group: %+v", group)
	//	return err
	//}
	//envs := generateEnvs(ctx, ss, group, az.Name, az.Code, jvmRatio)

	//volumes, volumeMounts, pvcs, err := createVolumes(ss, vols, group)
	//if err != nil {
	//	return err
	//}

	var imagePullSecretName []apiv1.LocalObjectReference

	if viper.GetBool(fmt.Sprintf("k8s.%s.container.imagePullSecret.enable", appDs.ZoneName)) {
		imagePullSecretName = []apiv1.LocalObjectReference{{
			Name: viper.GetString(fmt.Sprintf("k8s.%s.container.imagePullSecret.name", appDs.ZoneName))}}
	}

	//args, err := argsJson(appDs.Args)
	//if err != nil {
	//	log.WithError(err).Error("Failed marshal arguments to json string")
	//	return err
	//}

	dts := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        appDs.Name,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(ds.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"appName": appDs.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"appName":                appDs.Name,
						"cloud.graviti.cn/type":  "app",
						"cloud.graviti.cn/env":   appDs.Env,
						"cloud.graviti.cn/appid": appDs.AppID,
					},
					Annotations: map[string]string{
						podAnnoType:  annotations[podAnnoType],
						podAnnoEnv:   annotations[podAnnoEnv],
						podAnnoOrgID: annotations[podAnnoOrgID],
						podAnnoAppID: annotations[podAnnoAppID],
					},
				},
				Spec: apiv1.PodSpec{
					SecurityContext: &apiv1.PodSecurityContext{
						FSGroup: int64Ptr(2000),
					},
					//Volumes:          volumes,
					ImagePullSecrets: imagePullSecretName,

					Containers: []apiv1.Container{
						{
							Name:  appDs.Name,
							Image: ds.DsSpec.Image,
							//Command: []string{cinitEntrypoint},
							//Args: []string{
							//	"-logdir", "/mnt/mesos/sandbox",
							//	"-stdout", "/mnt/mesos/sandbox/stdout",
							//	"-stderr", "/mnt/mesos/sandbox/stderr",
							//	"-cmd", ss.Command,
							//	"-args", args},
							//Env:     envs,
							//EnvFrom: envFromSource(),
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(limitCPU(ds.DsSpec.CPU, appDs.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(limitMem(ds.DsSpec.Mem, appDs.ZoneName), apiresource.BinarySI),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(requestCPU(ds.DsSpec.CPU, appDs.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(requestMem(ds.DsSpec.Mem, appDs.ZoneName), apiresource.BinarySI),
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Capabilities: &apiv1.Capabilities{
									Add: []apiv1.Capability{"SYS_ADMIN"},
								},
							},
							//VolumeMounts: volumeMounts,
						},
					},
					DNSPolicy:   apiv1.DNSDefault,
					Affinity:    scheduleAffinity(appDs),
					Tolerations: scheduleToleration(appDs),
				},
			},
			//VolumeClaimTemplates: pvcs,
		},
	}

	if ds.DsSpec.HealthCheck == "" {
		log.Info("Skip setup health check")
	} else {
		log.Info("Setup health-check")
		url, err := url.ParseRequestURI(ds.DsSpec.HealthCheck)
		if err != nil {
			SendResponse(ctx, err, nil)
		}

		var port int64
		if port, err = strconv.ParseInt(strings.SplitN(url.Host, ":", 2)[1], 10, 32); err != nil {
			log.WithError(err).Errorf("Failed get url port")
			SendResponse(ctx, err, nil)
		}
		dts.Spec.Template.Spec.Containers[0].ReadinessProbe = &apiv1.Probe{
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(port),
					},
					Path:        url.Path,
					HTTPHeaders: []apiv1.HTTPHeader{{Name: "Accept", Value: "*/*"}},
				},
			},
			InitialDelaySeconds: 2,
			PeriodSeconds:       3,
			TimeoutSeconds:      2,
		}
	}

	begin := time.Now()
	dts, err = cs.AppsV1().Deployments(appDs.Namespace).Create(dts)
	if err != nil {
		SendResponse(ctx, err, "create deployment fail.")
		return
	}
	logger.MetricsEmit(
		"k8s.create_dps",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == err,
	)
	SendResponse(ctx, errno.OK, fmt.Sprintf("Create deployment %s success.", appDs.Name))
	return
}

// DeleteDeployment delete deployment instance
func DeleteDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog

	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	cs, err := GetClientByAzCode(zoneName)
	if err != nil {
		SendResponse(ctx, errno.ErrBind, nil)
		return
	}

	log.Info("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground

	begin := time.Now()
	err = cs.AppsV1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	logger.MetricsEmit(
		"k8s.delete_dts",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == nil || errors.IsNotFound(err),
	)
	if errors.IsNotFound(err) {
		log.Infof("Statefulset %s not found in k8s", name)
		SendResponse(ctx, err, nil)
	}
	if err != nil {
		SendResponse(ctx, err, nil)
	}

	SendResponse(ctx, errno.OK, nil)
}

// ScaleDeployment scale num of deployments
func ScaleDeployment(ctx *gin.Context) {

	var scaleDst *model.DeploymentScale
	if err := ctx.BindJSON(&scaleDst); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
		return
	}
	scaleApp := scaleDst.AppMeta
	cs, err := GetClientByAzCode(scaleApp.ZoneName)
	if err != nil {
		SendResponse(ctx, err, nil)
		return
	}

	begin := time.Now()
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := cs.AppsV1().Deployments(scaleApp.Namespace).Get(scaleApp.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		*result.Spec.Replicas = scaleDst.Replicas
		_, err = cs.AppsV1().Deployments(scaleApp.Namespace).Update(result)
		return err
	})
	logger.MetricsEmit(
		"k8s.scale_sts",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		retryErr == nil,
	)

	SendResponse(ctx, retryErr, nil)
}

// UpdateDeployment update deployment instance exisited
func UpdateDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	var upDst *model.Deployment
	if err := ctx.BindJSON(&upDst); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
		return
	}

	upApp := upDst.AppMeta
	cs, err := GetClientByAzCode(upApp.ZoneName)
	if err != nil {
		SendResponse(ctx, err, "nil")
		return
	}

	log.Info("Updating Deployment...")

	//volumes, volumeMounts, _, err := createVolumes(ss, vols, group)
	//if err != nil {
	//	return err
	//}

	//args, err := argsJson(ss.Args)
	//if err != nil {
	//	log.WithError(err).Error("Failed marshal arguments to json string")
	//	return err
	//}

	begin := time.Now()
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := cs.AppsV1().Deployments(upApp.Namespace).Get(upApp.Name, metav1.GetOptions{})

		// If Deployment not exists, we just skip update it
		if errors.IsNotFound(err) {
			log.Warnf("Deployment %s not exists in k8s cluster", upApp.Name)
			SendResponse(ctx, err, "nil")
		}
		if err != nil {
			return err
		}

		for _, anno := range upDst.DsSpec.Annotations {
			result.ObjectMeta.Annotations[anno.Key] = anno.Value
		}

		result.Spec.Template.ObjectMeta.Annotations[podAnnoType] = result.ObjectMeta.Annotations[podAnnoType]
		result.Spec.Template.ObjectMeta.Annotations[podAnnoEnv] = result.ObjectMeta.Annotations[podAnnoEnv]
		result.Spec.Template.ObjectMeta.Annotations[podAnnoOrgID] = result.ObjectMeta.Annotations[podAnnoOrgID]
		result.Spec.Template.ObjectMeta.Annotations[podAnnoAppID] = result.ObjectMeta.Annotations[podAnnoAppID]

		if viper.GetBool(fmt.Sprintf("k8s.%s.container.imagePullSecret.enable", upApp.ZoneName)) {
			result.Spec.Template.Spec.ImagePullSecrets = []apiv1.LocalObjectReference{{
				Name: viper.GetString(fmt.Sprintf("k8s.%s.container.imagePullSecret.name", upApp.ZoneName))}}
		}

		//result.Spec.Template.Spec.Volumes = volumes
		result.Spec.Template.Spec.Affinity = scheduleAffinity(upApp)
		result.Spec.Template.Spec.Tolerations = scheduleToleration(upApp)
		container := &result.Spec.Template.Spec.Containers[0]
		container.Image = upDst.DsSpec.Image
		//container.Command = []string{cinitEntrypoint}
		//container.Args = []string{
		//	"-logdir", "/mnt/mesos/sandbox",
		//	"-stdout", "/mnt/mesos/sandbox/stdout",
		//	"-stderr", "/mnt/mesos/sandbox/stderr",
		//	"-cmd", ss.Command,
		//	"-args", args}
		//container.Env = generateEnvs(ctx, ss, group, az.Name, az.Code, jvmRatio)
		//container.EnvFrom = envFromSource()
		//container.VolumeMounts = volumeMounts
		container.SecurityContext = &apiv1.SecurityContext{
			Capabilities: &apiv1.Capabilities{
				Add: []apiv1.Capability{"SYS_ADMIN"},
			},
		}

		container.Resources = apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(limitCPU(upDst.DsSpec.CPU, upApp.ZoneName), apiresource.DecimalSI),
				apiv1.ResourceMemory: *apiresource.NewQuantity(limitMem(upDst.DsSpec.Mem, upApp.ZoneName), apiresource.BinarySI),
			},
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(requestCPU(upDst.DsSpec.CPU, upApp.ZoneName), apiresource.DecimalSI),
				apiv1.ResourceMemory: *apiresource.NewQuantity(requestMem(upDst.DsSpec.Mem, upApp.ZoneName), apiresource.BinarySI),
			},
		}

		if upDst.DsSpec.HealthCheck == "" {
			container.ReadinessProbe = nil
		} else {
			url, err := url.ParseRequestURI(upDst.DsSpec.HealthCheck)
			if err != nil {
				return err
			}

			var port int64
			if port, err = strconv.ParseInt(strings.SplitN(url.Host, ":", 2)[1], 10, 32); err != nil {
				log.WithError(err).Errorf("Failed get url port")
				return err
			}
			container.ReadinessProbe = &apiv1.Probe{
				Handler: apiv1.Handler{
					HTTPGet: &apiv1.HTTPGetAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: int32(port),
						},
						Path:        url.Path,
						HTTPHeaders: []apiv1.HTTPHeader{{Name: "Accept", Value: "*/*"}},
					},
				},
				InitialDelaySeconds: 2,
				PeriodSeconds:       3,
				TimeoutSeconds:      2,
			}
		}
		_, err = cs.AppsV1().Deployments(upApp.Namespace).Update(result)
		return err
	})
	logger.MetricsEmit(
		"k8s.update_dts",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		retryErr == err,
	)
	if retryErr != nil {
		SendResponse(ctx, retryErr, "nil")
		return
	}
	SendResponse(ctx, errno.OK, nil)
}

func requestCPU(cpu float64, zoneName string) int64 {
	return int64(cpu * 1000 / viper.GetFloat64(fmt.Sprintf("k8s.%s.resource.cpuOverCommitRate", zoneName)))
}

func limitCPU(cpu float64, zoneName string) int64 {

	return int64(cpu * 1000 * viper.GetFloat64(fmt.Sprintf("k8s.%s.resource.cpuBurstRate", zoneName)))
}

func requestMem(mem int64, zoneName string) int64 {

	return int64(float64(mem*1024*1024) / viper.GetFloat64(fmt.Sprintf("k8s.%s.resource.memOverCommitRate", zoneName)))
}

func limitMem(mem int64, zoneName string) int64 {

	return int64(float64(mem*1024*1024) * viper.GetFloat64(fmt.Sprintf("k8s.%s.resource.memBurstRate", zoneName)))
}
