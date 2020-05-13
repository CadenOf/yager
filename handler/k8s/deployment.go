package k8s

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	. "voyager/handler"
	"voyager/model"
	DEP_CONST "voyager/pkg/constvar"
	"voyager/pkg/errno"
	"voyager/pkg/logger"
	"voyager/util"

	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetDeployment get deployment instances from k8s
func GetDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
		return
	}

	startAt := time.Now()
	dps, err := kclient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get deployment info.")
		return
	}
	logger.MetricsEmit(
		"k8s.get_dps",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)
	SendResponse(ctx, errno.OK, dps)
}

//ListDeployment List all deployments from specify namespace
func ListDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	startAt := time.Now()
	dep, err := kclient.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get deployment info.")
		return
	}
	logger.MetricsEmit(
		"k8s.get_dep",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)

	SendResponse(ctx, errno.OK, dep.Items)
}

// CreateDeployment create deployment instance
func CreateDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	var depModel *model.Deployment
	if err := ctx.BindJSON(&depModel); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
	}

	depNamespace := strings.ToLower(depModel.DepMeta.AppMeta.Namespace)
	depZone := depModel.DepMeta.AppMeta.ZoneName
	depName := depModel.DepMeta.AppMeta.Name

	// fetch k8s-client hander by zoneName
	kclient, err := GetClientByAzCode(depZone)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
		return
	}

	startAt := time.Now() // used to record operation time cost
	_, err = kclient.AppsV1().Deployments(depNamespace).Create(makeupDeploymentData(ctx, depModel))
	if err != nil {
		SendResponse(ctx, err, "create deployment fail.")
		return
	}
	logger.MetricsEmit(
		"k8s.create_dep",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)
	SendResponse(ctx, errno.OK, fmt.Sprintf("Create deployment %s success.", depName))
}

// DeleteDeployment delete deployment instance
func DeleteDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		SendResponse(ctx, errno.ErrBind, nil)
		return
	}

	log.Info("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground

	startAt := time.Now()
	err = kclient.AppsV1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	logger.MetricsEmit(
		"k8s.delete_dts",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
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

// ScaleDeployment scale num of deployment replicaset
func ScaleDeployment(ctx *gin.Context) {

	var scaleDep *model.DeploymentScale
	if err := ctx.BindJSON(&scaleDep); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
		return
	}
	scaleMeta := scaleDep.AppMeta
	kclient, err := GetClientByAzCode(scaleMeta.ZoneName)
	if err != nil {
		SendResponse(ctx, err, nil)
		return
	}

	startAt := time.Now()
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// fetch deployment template data by its namespace & deploymentName
		result, err := kclient.AppsV1().Deployments(scaleMeta.Namespace).Get(scaleMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		*result.Spec.Replicas = scaleDep.Replicas
		_, err = kclient.AppsV1().Deployments(scaleMeta.Namespace).Update(result)
		return err
	})
	logger.MetricsEmit(
		"k8s.scale_sts",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		retryErr == nil,
	)

	SendResponse(ctx, retryErr, nil)
}

// UpdateDeployment update deployment instance exisited
func UpdateDeployment(ctx *gin.Context) {
	log := logger.RuntimeLog
	var depModel *model.Deployment
	if err := ctx.BindJSON(&depModel); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
		return
	}

	depNamespace := depModel.DepMeta.AppMeta.Namespace
	depName := depModel.DepMeta.AppMeta.Name
	zoneName := depModel.DepMeta.AppMeta.ZoneName

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		SendResponse(ctx, err, "nil")
		return
	}

	log.Info("Updating Deployment...")
	startAt := time.Now()
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// fetch deployment template data by its namespace & deploymentName
		_, err := kclient.AppsV1().Deployments(depNamespace).Get(depName, metav1.GetOptions{})

		// If Deployment not exists, we just skip update it
		if errors.IsNotFound(err) {
			log.Warnf("Deployment %s not exists in k8s cluster", depName)
			SendResponse(ctx, err, "nil")
		}
		if err != nil {
			return err
		}

		// update deployment
		_, err = kclient.AppsV1().Deployments(depNamespace).Update(makeupDeploymentData(ctx, depModel))
		return err
	})

	logger.MetricsEmit(
		"k8s.update_dep",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		retryErr == err,
	)
	if retryErr != nil {
		SendResponse(ctx, retryErr, "nil")
		return
	}
	SendResponse(ctx, errno.OK, nil)
}

// Makeup Deployment TemplateData
func makeupDeploymentData(ctx *gin.Context, depModel *model.Deployment) *appsv1.Deployment {

	log := logger.RuntimeLog
	var affinity *model.AffinityStruct
	// var toleration *model.TolerationStruct
	depMeta := depModel.DepMeta.AppMeta
	depSpec := depModel.DepSpec.AppSpec

	affinity = &model.AffinityStruct{}
	affinity.AffMeta = depMeta
	affinity.Selector = depSpec.NodeSelector
	// toleration.TolerMeta = *depMeta
	// toleration.Toleration = depSpec.Toleration

	// init annotations
	annotations := map[string]string{
		DEP_CONST.K8S_RESOURCE_ANNOTATION_zone:  depMeta.ZoneName,
		DEP_CONST.K8S_RESOURCE_ANNOTATION_orgid: depMeta.OrgID,
		DEP_CONST.K8S_RESOURCE_ANNOTATION_appid: depMeta.AppID,
		DEP_CONST.K8S_RESOURCE_ANNOTATION_env:   depMeta.Env,
	}

	// init imagePullSecret
	var imagePullSecretName []apiv1.LocalObjectReference
	if viper.GetBool(fmt.Sprintf("k8s.%s.container.imagePullSecret.enable", depMeta.ZoneName)) {
		imagePullSecretName = []apiv1.LocalObjectReference{{
			Name: viper.GetString(fmt.Sprintf("k8s.%s.container.imagePullSecret.name", depMeta.ZoneName))}}
	}

	// init args
	//args, err := argsJson(appDs.Args)
	//if err != nil {
	//	log.WithError(err).Error("Failed marshal arguments to json string")
	//	return err
	//}

	// makeup overall subset
	depSet := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        depMeta.Name,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(depSpec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					DEP_CONST.K8S_RESOURCE_ANNOTATION_appid: depMeta.AppID,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						DEP_CONST.K8S_RESOURCE_ANNOTATION_env:   depMeta.Env,
						DEP_CONST.K8S_RESOURCE_ANNOTATION_appid: depMeta.AppID,
					},
				},
				Spec: apiv1.PodSpec{
					SecurityContext: &apiv1.PodSecurityContext{
						FSGroup: int64Ptr(2000),
					},
					ImagePullSecrets: imagePullSecretName,
					Containers: []apiv1.Container{
						{
							Name:  depMeta.Name,
							Image: depSpec.ContainerSpec.Image,
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
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(limitCPU(depSpec.ContainerSpec.CPU, depMeta.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(limitMem(depSpec.ContainerSpec.Mem, depMeta.ZoneName), apiresource.BinarySI),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(requestCPU(depSpec.ContainerSpec.CPU, depMeta.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(requestMem(depSpec.ContainerSpec.Mem, depMeta.ZoneName), apiresource.BinarySI),
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Capabilities: &apiv1.Capabilities{
									Add: []apiv1.Capability{"SYS_ADMIN"},
								},
							},
						},
					},
					DNSPolicy: apiv1.DNSDefault,
					Affinity:  scheduleAffinity(affinity),
					// Tolerations: scheduleToleration(toleration),
				},
			},
		},
	}

	// init healthCheck endpoint
	if depSpec.ContainerSpec.HealthCheck == "" {
		log.Info("Skip setup health check")
	} else {
		log.Info("Setup health-check")
		url, err := url.ParseRequestURI(depSpec.ContainerSpec.HealthCheck)
		if err != nil {
			SendResponse(ctx, err, nil)
		}

		var port int64
		if port, err = strconv.ParseInt(strings.SplitN(url.Host, ":", 2)[1], 10, 32); err != nil {
			log.WithError(err).Errorf("Failed get url port")
			SendResponse(ctx, err, nil)
		}
		depSet.Spec.Template.Spec.Containers[0].ReadinessProbe = &apiv1.Probe{
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
			InitialDelaySeconds: DEP_CONST.K8S_DEPLOYMENT_InitialDelaySeconds,
			PeriodSeconds:       DEP_CONST.K8S_DEPLOYMENT_PeriodSeconds,
			TimeoutSeconds:      DEP_CONST.K8S_DEPLOYMENT_TimeoutSeconds,
		}
	}
	return depSet
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
