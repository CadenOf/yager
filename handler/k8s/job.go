package k8s

import (
	"fmt"
	//"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"
	"voyager/model"

	. "voyager/handler"
	JOB_CONST "voyager/pkg/constvar"
	"voyager/pkg/errno"
	"voyager/pkg/logger"
	"voyager/util"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//GetJob get job instance from k8s
func GetJob(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	startAt := time.Now()
	jbs, err := kclient.BatchV1().Jobs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		SendResponse(ctx, err, "failed to get job info.")
		return
	}
	logger.MetricsEmit(
		"k8s.get_job",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)

	SendResponse(ctx, errno.OK, jbs)
}

// CreateJob create job instance in k8s
func CreateJob(ctx *gin.Context) {
	log := logger.RuntimeLog
	var jobModel *model.Job
	if err := ctx.BindJSON(&jobModel); err != nil {
		SendResponse(ctx, err, "Request Body Invalid")
	}

	jobNamespace := strings.ToLower(jobModel.JobMeta.AppMeta.Namespace)
	jobName := jobModel.JobMeta.AppMeta.Name
	zoneName := jobModel.JobMeta.AppMeta.ZoneName

	// fetch k8s-client handler by zoneName
	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		log.WithError(err)
		SendResponse(ctx, errno.ErrTokenInvalid, nil)
	}

	startAt := time.Now()
	_, err = kclient.BatchV1().Jobs(jobNamespace).Create(makeupJobData(ctx, jobModel))
	if err != nil {
		SendResponse(ctx, err, "create deployment fail.")
		return
	}
	logger.MetricsEmit(
		"k8s.create_jobs",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
		err == err,
	)
	SendResponse(ctx, errno.OK, fmt.Sprintf("Create Job %s success.", jobName))
	return
}

//DeleteJob delete job instance in k8s
func DeleteJob(ctx *gin.Context) {
	log := logger.RuntimeLog
	zoneName := ctx.Param("zone")
	namespace := ctx.Param("ns")
	name := ctx.Param("name")

	kclient, err := GetClientByAzCode(zoneName)
	if err != nil {
		SendResponse(ctx, errno.ErrBind, nil)
		return
	}

	log.Info("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground

	startAt := time.Now()
	err = kclient.BatchV1().Jobs(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	logger.MetricsEmit(
		"k8s.delete_jobs",
		util.GetReqID(ctx),
		float32(time.Since(startAt)/time.Millisecond),
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

func makeupJobData(ctx *gin.Context, jobModel *model.Job) *batchv1.Job {
	log := logger.RuntimeLog
	var affinity *model.AffinityStruct
	jobMeta := jobModel.JobMeta.AppMeta
	jobSpec := jobModel.JobSpec.AppSpec

	affinity.AffMeta = jobMeta
	affinity.Selector = jobSpec.NodeSelector
	// toleration.TolerMeta = jobMeta
	// toleration.Toleration = jobSpec.Toleration

	// init annotations
	annotations := map[string]string{
		JOB_CONST.K8S_RESOURCE_ANNOTATION_zone:  jobMeta.ZoneName,
		JOB_CONST.K8S_RESOURCE_ANNOTATION_orgid: jobMeta.OrgID,
		JOB_CONST.K8S_RESOURCE_ANNOTATION_appid: jobMeta.AppID,
		JOB_CONST.K8S_RESOURCE_ANNOTATION_env:   jobMeta.Env,
	}

	// init imagePullSecret
	var imagePullSecretName []apiv1.LocalObjectReference
	if viper.GetBool(fmt.Sprintf("k8s.%s.container.imagePullSecret.enable", jobMeta.ZoneName)) {
		imagePullSecretName = []apiv1.LocalObjectReference{{
			Name: viper.GetString(fmt.Sprintf("k8s.%s.container.imagePullSecret.name", jobMeta.ZoneName))}}
	}

	// init args
	//args, err := argsJson(jbs.Args)
	//if err != nil {
	//	log.WithError(err).Error("Failed marshal arguments to json string")
	//	return err
	//}

	// init overall subset
	jobSet := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobMeta.Name,
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
						JOB_CONST.K8S_RESOURCE_ANNOTATION_appid: jobMeta.AppID,
						JOB_CONST.K8S_RESOURCE_ANNOTATION_env:   jobMeta.Env,
						JOB_CONST.K8S_RESOURCE_ANNOTATION_zone:  jobMeta.ZoneName,
					},
					Annotations: annotations,
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
							Name:  jobMeta.Name,
							Image: jobSpec.ContainerSpec.Image,
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
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(limitCPU(jobSpec.ContainerSpec.CPU, jobMeta.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(limitMem(jobSpec.ContainerSpec.Mem, jobMeta.ZoneName), apiresource.BinarySI),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    *apiresource.NewMilliQuantity(requestCPU(jobSpec.ContainerSpec.CPU, jobMeta.ZoneName), apiresource.DecimalSI),
									apiv1.ResourceMemory: *apiresource.NewQuantity(requestMem(jobSpec.ContainerSpec.Mem, jobMeta.ZoneName), apiresource.BinarySI),
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
					DNSPolicy: apiv1.DNSDefault,
					Affinity:  scheduleAffinity(affinity),
					// Tolerations: scheduleToleration(toleration),
				},
			},
			//VolumeClaimTemplates: pvcs,
		},
	}

	// init healthCheck endpoint
	if jobSpec.ContainerSpec.HealthCheck == "" {
		log.Info("Skip setup health check")
	} else {
		log.Info("Setup health-check")
		url, err := url.ParseRequestURI(jobSpec.ContainerSpec.HealthCheck)
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
			InitialDelaySeconds: JOB_CONST.K8S_JOB_InitialDelaySeconds,
			PeriodSeconds:       JOB_CONST.K8S_JOB_PeriodSeconds,
			TimeoutSeconds:      JOB_CONST.K8S_JOBT_TimeoutSeconds,
		}
	}

	return jobSet
}
