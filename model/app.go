package model

import (
	apiv1 "k8s.io/api/core/v1"
)

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
	Replicas      int32         `json:"replicas" binding:"required" `
	Annotations   []KV          `xorm:"TEXT json"`
	Tolerations   []KV          `json:"tolerations"`
	NodeSelector  []KVL         `json:"nodeSelector" binding:"required`
	ContainerSpec ContainerInfo `json:"containerSpec"`
}

type ContainerInfo struct {
	CPU         float64             `xorm:"name cpu" binding:"required"`
	Mem         int64               `json:"mem" binding:"required"`
	DiskSize    int64               `json:"diskSize"`
	Command     string              `json:"command"`
	Args        []string            `json:"args"`
	Image       string              `json:"image" binding:"required"`
	Envs        []KV                `json:"envs"`
	HealthCheck string              `json:"healthCheck"`
	Volumes     []apiv1.VolumeMount `json:"volumes,omitempty"`
}

// common struct
type AffinityInfo struct {
	AffMeta  AppMetaInfo
	Selector []KVL `json:"nodeSelector"`
}

type TolerationInfo struct {
	TolerMeta  AppMetaInfo
	Toleration []KV `xorm:"TEXT json"`
}
