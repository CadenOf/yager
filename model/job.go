package model

type Job struct {
	AppMeta           *App
	Replicas         int32  `json:"replicas"`
	JobSpec           JobSpec
	//vols []model.Volume
}

type JobSpec struct {
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

type JobScale struct {
	AppMeta           *App
	Replicas         int32  `json:"replicas" binding:"required"`
}



