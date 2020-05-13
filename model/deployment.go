package model

type Deployment struct {
	DepMeta DeploymentMeta
	DepSpec DeploymentSpec
}

type DeploymentSpec struct {
	AppSpec       AppSpecInfo
	WarmUpTimeout int64 `json:"warmUpTimeout"`
}

type DeploymentMeta struct {
	AppMeta AppMetaInfo
}

type DeploymentScale struct {
	AppMeta  *AppMetaInfo
	Replicas int32 `json:"replicas" binding:"required"`
}

type DeploymentFull struct {
	Ds   Deployment
	Spec *DeploymentSpec `json:",omitempty"`
	Pods []*Pod
}
