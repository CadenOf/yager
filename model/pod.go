package model

type FailAndRecoveryCount struct {
	FailCount     int64
	RecoveryCount int64
}

type Pod struct {
	Id                       int64
	FQDN                     string `xorm:"name fqdn"`
	Name                     string `json:"-"`
	IP                       string `xorm:"name ip"`
	PublicIP                 string `xorm:"name public_ip"`
	Order                    int64
	Namespace                string
	SetId                    int64
	GroupId                  int64
	AzCode                   string `json:"-"`
	AzsetId                  int64
	SpecId                   int64
	CICode                   string `xorm:"name cicode"`
	IsDeleting               bool
	IsDeploy                 bool `xorm:"name is_deploy"`
	Platform                 string
	AppId                    string
	PortId                   string
	TsAndVersion             `xorm:"extends"`
	FailAndRecoveryCount     `xorm:"extends"`
}

type PodBatch struct {
	PodsFQDN []string
}
