package constvar

const (
	// deployment constvars
	K8S_DEPLOYMENT_InitialDelaySeconds = 2
	K8S_DEPLOYMENT_PeriodSeconds       = 3
	K8S_DEPLOYMENT_TimeoutSeconds      = 3

	K8S_VOLUME_Name              = "app-log"
	K8S_VOLUME_RootPath          = "/mnt/k8s/"
	K8S_VOLUME_SubPath           = "logs"
	K8S_VOLUME_ContainerRootPath = "/var/log/k8s/"

	K8S_ENV_PAAS_ZONE          = "PAAS_ZONE"
	K8S_ENV_PAAS_APPID         = "PAAS_APPID"
	K8S_ENV_PAAS_POD_NAME      = "PAAS_POD_NAME"
	K8S_ENV_PAAS_HYPERVISOR    = "PAAS_HYPERVISOR"
	K8S_ENV_PAAS_HYPERVISOR_IP = "PAAS_HYPERVISOR_IP"
	K8S_ENV_PAAS_POD_NAMESPACE = "PAAS_POD_NAMESPACE"
	K8S_ENV_PAAS_POD_SA        = "PAAS_POD_SA"
	K8S_ENV_PAAS_POD_IP        = "PAAS_POD_IP"

	// job constvars
	K8S_JOB_InitialDelaySeconds = 2
	K8S_JOB_PeriodSeconds       = 3
	K8S_JOBT_TimeoutSeconds     = 3

	// annotations key
	K8S_RESOURCE_ANNOTATION_zone  = "paas.graviti.cn/zone"
	K8S_RESOURCE_ANNOTATION_orgid = "paas.graviti.cn/orgid"
	K8S_RESOURCE_ANNOTATION_appid = "paas.graviti.cn/appid"
	K8S_RESOURCE_ANNOTATION_env   = "paas.graviti.cn/env"
	K8S_RESOURCE_ANNOTATION_type  = "paas.graviti.cn/type"

	// log msg record
	K8S_LOG_Method_GetDeployment    = "k8s.get_depoyment"
	K8S_LOG_Method_ListDeployment   = "k8s.list_deployment"
	K8S_LOG_Method_CreateDeployment = "k8s.create_deployment"
	K8S_LOG_Method_DeleteDeployment = "k8s.delete_deployment"
	K8S_LOG_Method_UpdateDeployment = "k8s.update_deployment"
	K8S_LOG_Method_ScaleDeployment  = "k8s.scale_deployment"

	K8S_LOG_Method_GetJob    = "k8s.get_job"
	K8S_LOG_Method_ListJob   = "k8s.list_job"
	K8S_LOG_Method_CreateJob = "k8s.create_job"
	K8S_LOG_Method_DeleteJob = "k8s.delete_job"

	K8S_LOG_Method_GetService    = "k8s.get_service"
	K8S_LOG_Method_ListService   = "k8s.list_service"
	K8S_LOG_Method_CreateService = "k8s.create_service"
	K8S_LOG_Method_DeleteService = "k8s.delete_service"
)
