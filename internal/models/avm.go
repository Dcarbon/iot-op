package models

import (
	"time"

	"github.com/Dcarbon/go-shared/libs/decimal"
)

const (
	TableNameAVM = "avms"
)

type AVM struct {
	Id        string          `json:"id"         `                                       //
	Signed    string          `json:"signed"     `                                       // Base 64
	Data      string          `json:"data"       `                                       //
	Volume    decimal.Decimal `json:"vol"        gorm:"type:decimal"`                    //
	IotId     int64           `json:"iotId"      gorm:"index:avm_iot_ca_idx,priority:1"` //
	CreatedAt time.Time       `json:"createdAt"  gorm:"index:avm_iot_ca_idx,priority:2"` //
}

func (*AVM) TableName() string {
	return TableNameAVM
}
