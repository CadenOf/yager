package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"

	//"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"
	"voyager/model"

	. "voyager/handler"
	"voyager/pkg/errno"
	"voyager/pkg/logger"
	"voyager/util"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
)

//GetJob get job instance from k8s
func GetJob(ctx *gin.Context) {
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
	jbs, err := cs.BatchV1().Jobs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get job info.")
		return
	}
	logger.MetricsEmit(
		"k8s.get_job",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == err,
	)

	SendResponse(ctx, errno.OK, jbs)
}

// CreateJob create job instance in k8s
func CreateJob(ctx *gin.Context) {

	log := logger.RuntimeLog

	var jbs *model.Job
	if err := ctx.BindJSON(&jbs); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
	}
	appJob := jbs.AppMeta

	fmt.Printf("name: %s %s %s \n", appJob.Name, appJob.ZoneName, appJob.AppID)
	appJob.Namespace = fmt.Sprintf("%s-app", strings.ToLower(appJob.Env))

	cs, err := GetClientByAzCode(appJob.ZoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	annotations := map[string]string{
		"cloud.graviti.cn/zone":  appJob.ZoneName,
		"cloud.graviti.cn/orgid": appJob.OrgID,
		"cloud.graviti.cn/appid": appJob.AppID,
		"cloud.graviti.cn/env":   appJob.Env,
	}

	for _, anno := range jbs.JobSpec.Annotations {
		annotations[anno.Key] = anno.Value
	}

	var imagePullSecretName []apiv1.LocalObjectReference

	if viper.GetBool(fmt.Sprintf("k8s.%s.container.imagePullSecret.enable", appJob.ZoneName)) {
		imagePullSecretName = []apiv1.LocalObjectReference{{
			Name: viper.GetString(fmt.Sprintf("k8s.%s.container.imagePullSecret.name", appJob.ZoneName))}}
	}

	//args, err := argsJson(jbs.Args)
	//if err != nil {
	//	log.WithError(err).Error("Failed marshal arguments to json string")
	//	return err
	//}

	jobSet := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        appJob.Name,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			//Selector: &metav1.LabelSelector{
			//	MatchLabels: map[string]string{
			//		"appName": appJob.Name,
			//	},
			//	MatchExpressions: []LabelSelectorRequirement{
			//
			//	},
			//},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"appName":                appJob.Name,
						"cloud.graviti.cn/type":  "app",
						"cloud.graviti.cn/env":   appJob.Env,
						"cloud.graviti.cn/appid": appJob.AppID,
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
					RestartPolicy:    "Never",
					Containers: []apiv1.Container{
						{
							Name:  appJob.Name,
							Image: jbs.JobSpec.Image,
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
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(limitCPU(jbs.JobSpec.CPU, appJob.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(limitMem(jbs.JobSpec.Mem, appJob.ZoneName), apiresource.BinarySI),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(requestCPU(jbs.JobSpec.CPU, appJob.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(requestMem(jbs.JobSpec.Mem, appJob.ZoneName), apiresource.BinarySI),
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
					Affinity:    scheduleAffinity(appJob),
					Tolerations: scheduleToleration(appJob),
				},
			},
			//VolumeClaimTemplates: pvcs,
		},
	}

	if jbs.JobSpec.HealthCheck == "" {
		log.Info("Skip setup health check")
	} else {
		log.Info("Setup health-check")
		url, err := url.ParseRequestURI(jbs.JobSpec.HealthCheck)
		if err != nil {
			SendResponse(ctx, err, nil)
		}

		var port int64
		if port, err = strconv.ParseInt(strings.SplitN(url.Host, ":", 2)[1], 10, 32); err != nil {
			log.WithError(err).Errorf("Failed get url port")
			SendResponse(ctx, err, nil)
		}
		jobSet.Spec.Template.Spec.Containers[0].ReadinessProbe = &apiv1.Probe{
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
	jobSet, err = cs.BatchV1().Jobs(appJob.Namespace).Create(jobSet)
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
	SendResponse(ctx, errno.OK, fmt.Sprintf("Create deployment %s success.", appJob.Name))
	return
}

//DeleteJob delete job instance in k8s
func DeleteJob(ctx *gin.Context) {
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
	err = cs.BatchV1().Jobs(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	logger.MetricsEmit(
		"k8s.delete_jobs",
		util.GetReqID(ctx),
		float32(time.Since(begin)/time.Millisecond),
		err == nil || errors.IsNotFound(err),
	)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("Job %s not found in k8s", name)
			SendResponse(ctx, err, nil)
			return
		}
		SendResponse(ctx, err, nil)
		return
	}

	SendResponse(ctx, errno.OK, nil)
	return
}

//func ScaleJob(ctx *gin.Context) {
//
//	var scaleDst *model.DeploymentScale
//	if err := ctx.BindJSON(&scaleDst); err != nil {
//		SendResponse(ctx, err, "Request Body Invalid")
//		return
//	}
//
//	cs, err := GetClientByAzCode(scaleDst.ZoneName)
//	if err != nil {
//		SendResponse(ctx, err, nil)
//		return
//	}
//
//	begin := time.Now()
//	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
//		result, err := cs.AppsV1().Deployments(scaleDst.Namespace).Get(scaleDst.Name, metav1.GetOptions{})
//		if err != nil {
//			return err
//		}
//
//		*result.Spec.Replicas = scaleDst.Replicas
//		_, err = cs.AppsV1().Deployments(scaleDst.Namespace).Update(result)
//		return err
//	})
//	logger.MetricsEmit(
//		"k8s.scale_sts",
//		util.GetReqID(ctx),
//		float32(time.Since(begin)/time.Millisecond),
//		retryErr == nil,
//	)
//
//	SendResponse(ctx, retryErr, nil)
//}
//
//func UpdateJob(ctx *gin.Context) {
//	log := logger.RuntimeLog
//	var upDst *model.Deployment
//	if err := ctx.BindJSON(&upDst); err != nil {
//		SendResponse(ctx, err, "Request Body Invalid")
//		return
//	}
//
//	cs, err := GetClientByAzCode(upDst.ZoneName)
//	if err != nil {
//		SendResponse(ctx, err, "nil")
//		return
//	}
//
//	log.Info("Updating Deployment...")
//
//	//volumes, volumeMounts, _, err := createVolumes(ss, vols, group)
//	//if err != nil {
//	//	return err
//	//}
//
//	//args, err := argsJson(ss.Args)
//	//if err != nil {
//	//	log.WithError(err).Error("Failed marshal arguments to json string")
//	//	return err
//	//}
//
//	begin := time.Now()
//	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
//		result, err := cs.AppsV1().Deployments(upDst.Namespace).Get(upDst.Name, metav1.GetOptions{})
//
//		// If Deployment not exists, we just skip update it
//		if errors.IsNotFound(err) {
//			log.Warnf("Deployment %s not exists in k8s cluster", upDst.Name)
//			SendResponse(ctx, err, "nil")
//		}
//		if err != nil {
//			return err
//		}
//
//		for _, anno := range upDst.DsSpec.Annotations {
//			result.ObjectMeta.Annotations[anno.Key] = anno.Value
//		}
//
//		result.Spec.Template.ObjectMeta.Annotations[podAnnoType] = result.ObjectMeta.Annotations[podAnnoType]
//		result.Spec.Template.ObjectMeta.Annotations[podAnnoEnv] = result.ObjectMeta.Annotations[podAnnoEnv]
//		result.Spec.Template.ObjectMeta.Annotations[podAnnoOrgId] = result.ObjectMeta.Annotations[podAnnoOrgId]
//		result.Spec.Template.ObjectMeta.Annotations[podAnnoAppId] = result.ObjectMeta.Annotations[podAnnoAppId]
//
//		if viper.GetBool(fmt.Sprintf("k8s.%s.container.imagePullSecret.enable", upDst.ZoneName)) {
//			result.Spec.Template.Spec.ImagePullSecrets = []apiv1.LocalObjectReference{{
//				Name: viper.GetString(fmt.Sprintf("k8s.%s.container.imagePullSecret.name", upDst.ZoneName))}}
//		}
//
//		//result.Spec.Template.Spec.Volumes = volumes
//		result.Spec.Template.Spec.Affinity = scheduleAffinity(upDst)
//		result.Spec.Template.Spec.Tolerations = scheduleToleration(upDst)
//		container := &result.Spec.Template.Spec.Containers[0]
//		container.Image = upDst.DsSpec.Image
//		//container.Command = []string{cinitEntrypoint}
//		//container.Args = []string{
//		//	"-logdir", "/mnt/mesos/sandbox",
//		//	"-stdout", "/mnt/mesos/sandbox/stdout",
//		//	"-stderr", "/mnt/mesos/sandbox/stderr",
//		//	"-cmd", ss.Command,
//		//	"-args", args}
//		//container.Env = generateEnvs(ctx, ss, group, az.Name, az.Code, jvmRatio)
//		//container.EnvFrom = envFromSource()
//		//container.VolumeMounts = volumeMounts
//		container.SecurityContext = &apiv1.SecurityContext{
//			Capabilities: &apiv1.Capabilities{
//				Add: []apiv1.Capability{"SYS_ADMIN"},
//			},
//		}
//
//		container.Resources = apiv1.ResourceRequirements{
//			Limits: apiv1.ResourceList{
//				apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(limitCPU(upDst.DsSpec.CPU, upDst.ZoneName), apiresource.DecimalSI),
//				apiv1.ResourceMemory: *apiresource.NewQuantity(limitMem(upDst.DsSpec.Mem, upDst.ZoneName), apiresource.BinarySI),
//			},
//			Requests: apiv1.ResourceList{
//				apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(requestCPU(upDst.DsSpec.CPU, upDst.ZoneName), apiresource.DecimalSI),
//				apiv1.ResourceMemory: *apiresource.NewQuantity(requestMem(upDst.DsSpec.Mem, upDst.ZoneName), apiresource.BinarySI),
//			},
//		}
//
//		if upDst.DsSpec.HealthCheck == "" {
//			container.ReadinessProbe = nil
//		} else {
//			url, err := url.ParseRequestURI(upDst.DsSpec.HealthCheck)
//			if err != nil {
//				return err
//			}
//
//			var port int64
//			if port, err = strconv.ParseInt(strings.SplitN(url.Host, ":", 2)[1], 10, 32); err != nil {
//				log.WithError(err).Errorf("Failed get url port")
//				return err
//			}
//			container.ReadinessProbe = &apiv1.Probe{
//				Handler: apiv1.Handler{
//					HTTPGet: &apiv1.HTTPGetAction{
//						Port: intstr.IntOrString{
//							Type:   intstr.Int,
//							IntVal: int32(port),
//						},
//						Path:        url.Path,
//						HTTPHeaders: []apiv1.HTTPHeader{{Name: "Accept", Value: "*/*"}},
//					},
//				},
//				InitialDelaySeconds: 2,
//				PeriodSeconds:       3,
//				TimeoutSeconds:      2,
//			}
//		}
//		_, err = cs.AppsV1().Deployments(upDst.Namespace).Update(result)
//		return err
//	})
//	logger.MetricsEmit(
//		"k8s.update_dts",
//		util.GetReqID(ctx),
//		float32(time.Since(begin)/time.Millisecond),
//		retryErr == err,
//	)
//	if retryErr != nil{
//		SendResponse(ctx, retryErr,"nil")
//		return
//	}
//	SendResponse(ctx, errno.OK, nil)
//}
