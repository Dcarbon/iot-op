package domain

import (
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/iot-op/internal/models"
)

type RMinterMint struct {
	Nonce  int64  `json:"nonce" binding:"required"`  //
	Amount string `json:"amount" binding:"required"` // Hex
	Iot    string `json:"iot" binding:"required"`    // IoT Address
	Signed string `json:"signed" binding:"required"` // Base 64 (RSV string )
} // @name RMinter

type RMinterGetSigns struct {
	From  int64        `json:"from" form:"from" binding:"required"`
	To    int64        `json:"to" form:"to" binding:""`
	IotId int64        `json:"iotId" uri:"iotId" binding:"required"`
	Skip  int          `json:"skip" form:"skip"`
	Limit int          `json:"limit" form:"limit"`
	Sort  dmodels.Sort `json:"sort" form:"sort"`
} //@name RMinterGetList

type RMinterGetSignLatest struct {
	IotId int64 `json:"iotId" uri:"iotId" binding:"required"`
} //@name RMinterGetList

type RMinterGetMinted struct {
	From     int64             `json:"from" form:"from" binding:"required"` //
	To       int64             `json:"to" form:"to" binding:"required"`     //
	IotId    int64             `json:"iotId" form:"iotId" binding:""`       //
	Skip     int               `json:"skip" form:"skip"`                    //
	Limit    int64             `json:"limit"`                               //
	Sort     dmodels.Sort      `json:"sort" form:"sort"`                    //
	Interval dmodels.DInterval `json:"interval" form:"interval"`            // 1 : day 2: month
} //@name RMinterGetMinted

func (r *RMinterGetMinted) Normalize() {
	if r.To == 0 {
		r.To = time.Now().Unix()
		if r.From == 0 {
			r.From = time.Now().Unix() - 30*86400
		}
	}

	if r.Limit == 0 || r.Limit > 100 {
		r.Limit = 100
	}
}

type IMinter interface {
	Mint(*RMinterMint) error

	GetSigns(*RMinterGetSigns) ([]*models.MintSign, error)
	GetSignLatest(*RMinterGetSignLatest) (*models.MintSign, error)

	GetMinted(*RMinterGetMinted) ([]*models.Minted, error)

	GetSeparator() (*esign.TypedDataDomain, error)
}
