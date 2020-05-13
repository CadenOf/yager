package constvar

const (
	// deployment constvars
	K8S_DEPLOYMENT_InitialDelaySeconds = 2
	K8S_DEPLOYMENT_PeriodSeconds       = 3
	K8S_DEPLOYMENT_TimeoutSeconds      = 3

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
)
