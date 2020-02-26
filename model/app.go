package model

type App struct {
	Name             string `json:"name" binding:"required" validate:"min=1,max=128"`
	ZoneName         string `json:"zoneName" binding:"required" validate:"min=1,max=50"`
	Env              string `json:"env" binding:"required" validate:"min=1,max=50"`
	Namespace        string `json:"namespace"`
	OrgID            string `json:"orgid"`
	AppID            string `json:"appid" binding:"required"`
}