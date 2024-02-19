package domain

import (
	"encoding/base64"
	"encoding/json"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/libs/decimal"
	"github.com/Dcarbon/iot-op/internal/models"
)

type AVMSignExtract struct {
	Volume decimal.Decimal `json:"volume"`
	From   int64           `json:"from"`
	To     int64           `json:"to"`
}

type RAVMCreate struct {
	dmodels.Signature
	// AVMSignExtract
}

func (rCreate *RAVMCreate) Extract() (*AVMSignExtract, error) {
	raw, err := base64.StdEncoding.DecodeString(rCreate.Data)
	if nil != err {
		return nil, err
	}

	var rs = &AVMSignExtract{}
	err = json.Unmarshal(raw, rs)
	if nil != err {
		return nil, err
	}

	return rs, nil
}

type RAVMGetList struct {
	Full     bool              ``
	Skip     int               ``
	Limit    int               ``
	IotId    int64             ``
	From     int64             ``
	To       int64             ``
	Sort     dmodels.Sort      ``
	Interval dmodels.DInterval `json:"interval" form:"interval"` // 1 : day 2: month
}

type IAVM interface {
	Create(*RAVMCreate) (*models.AVM, error)
	GetList(*RAVMGetList) ([]*models.AVM, error)
}
