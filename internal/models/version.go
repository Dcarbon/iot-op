package models

import "time"

const (
	TableNameVersion = "iot_version"
)

type Version struct {
	IotType   int32     `json:"iot_type" gorm:"primary_key"`
	Version   string    `json:"version" gorm:"primary_key"`
	Path      string    `json:"path" gorm:"path"`
	CreatedAt time.Time `json:"createdAt" ` //
	UpdatedAt time.Time `json:"updatedAt" ` //

}

func (*Version) TableName() string { return TableNameVersion }
