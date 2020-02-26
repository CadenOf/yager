package model

type Deployment struct {
	AppMeta            *App
	Replicas         int32  `json:"replicas"`
	DsSpec           DeploymentSpec
	//vols []model.Volume
}

type DeploymentSpec struct {
	CPU              float64 `xorm:"name cpu"`
	Mem              int64
	DiskSize         int64
	Command          string
	Args             []string `xorm:"TEXT json"`
	HealthCheck      string `json:"healthCheck"`
	WarmUpTimeout    int64
	Image            string `json:"image"`
	Envs             []KV `xorm:"TEXT json"`
	Annotations      []KV `xorm:"TEXT json"`
}

type DeploymentScale struct {
	AppMeta            *App
	Replicas         int32  `json:"replicas" binding:"required"`
}

type DeploymentFull struct {
	Ds  Deployment
	Spec *DeploymentSpec `json:",omitempty"`
	Pods []*Pod
}


