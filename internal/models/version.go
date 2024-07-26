package models

const (
	TableNameVersion = "iot_version"
)

type Version struct {
	IotType int32  `json:"iot_type" gorm:"primary_key"`
	Version string `json:"version" gorm:"version"`
	Path    string `json:"path" gorm:"path"`
}

func (*Version) TableName() string { return TableNameVersion }
