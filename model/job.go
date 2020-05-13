package model

type Job struct {
	JobMeta JobMetaStruct
	JobSpec JobSpecStruct
}

type JobSpecStruct struct {
	AppSpec AppSpecInfo
}

type JobMetaStruct struct {
	AppMeta AppMetaInfo
}

type JobScale struct {
	JobMeta  AppMetaInfo
	Replicas int32 `json:"replicas" binding:"required"`
}
