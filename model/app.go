package model

// common struct

type AppMetaInfo struct {
	Name      string `json:"name" binding:"required" validate:"min=1,max=128"`
	ZoneName  string `json:"zoneName" binding:"required" validate:"min=1,max=50"`
	Env       string `json:"env" binding:"required" validate:"min=1,max=50"`
	Namespace string `json:"namespace" binding:"required"`
	OrgID     string `json:"orgid" binding:"required"`
	AppID     string `json:"appid" binding:"required"`
}

type AppSpecInfo struct {
	Replicas      int32 `json:"replicas" binding:"required" `
	Annotations   []KV  `xorm:"TEXT json"`
	Toleration    []KV  `xorm:"TEXT json"`
	NodeSelector  []KVL `xorm:"TEXT json" binding:"required`
	ContainerSpec ContainerInfo
}

type ContainerInfo struct {
	CPU         float64 `xorm:"name cpu"`
	Mem         int64
	DiskSize    int64
	Command     string
	Args        []string `xorm:"TEXT json"`
	Image       string   `json:"image" binding:"required"`
	Envs        []KV     `xorm:"TEXT json"`
	HealthCheck string   `json:"healthCheck"`
}

// common struct
type AffinityStruct struct {
	AffMeta  AppMetaInfo
	Selector []KVL `xorm:"TEXT json"`
}

type TolerationStruct struct {
	TolerMeta  AppMetaInfo
	Toleration []KV `xorm:"TEXT json"`
}
