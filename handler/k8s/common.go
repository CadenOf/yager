package k8s

import (
	"voyager/model"

	CONST "voyager/pkg/constvar"

	apiv1 "k8s.io/api/core/v1"
)

// Generate Envs for k8s resouces
func generateEnvs(appMeta *model.AppMetaInfo, appEnv []model.KV) []apiv1.EnvVar {

	// init envs
	envs := []apiv1.EnvVar{
		{Name: CONST.K8S_ENV_PAAS_ZONE, Value: appMeta.ZoneName},
		{Name: CONST.K8S_ENV_PAAS_APPID, Value: appMeta.AppID},
		{Name: CONST.K8S_ENV_PAAS_POD_NAME, ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
		{Name: CONST.K8S_ENV_PAAS_HYPERVISOR, ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
		{Name: CONST.K8S_ENV_PAAS_HYPERVISOR_IP, ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "status.hostIP"}}},
		{Name: CONST.K8S_ENV_PAAS_POD_NAMESPACE, ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		{Name: CONST.K8S_ENV_PAAS_POD_SA, ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}}},
		{Name: CONST.K8S_ENV_PAAS_POD_IP, ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "status.podIP"}}},
	}

	for _, e := range appEnv {
		envs = append(envs, apiv1.EnvVar{Name: e.Key, Value: e.Value})
	}

	return envs
}
