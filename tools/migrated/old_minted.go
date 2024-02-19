package main

import (
	"encoding/base64"
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/ecodes"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/iot-op/internal/models"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	OldTableNameMintSign = "mint_sign"
	OldTableNameMinted   = "minted"
)

type OldMintSign struct {
	ID        int64     `json:"id" gorm:"primary_key"` //
	IotId     int64     `json:"iotId"`                 // IoT id
	Nonce     int64     `json:"nonce" gorm:"index"`    //
	Amount    string    `json:"amount" `               // Hex
	Iot       string    `json:"iot" gorm:"index"`      // IoT Address
	R         string    `json:"r" `                    //
	S         string    `json:"s" `                    //
	V         string    `json:"v" `                    //
	CreatedAt time.Time `json:"createdAt" `            //
	UpdatedAt time.Time `json:"updatedAt" `            //
}

func (*OldMintSign) TableName() string { return OldTableNameMintSign }

// Only for test
// pk: private key (hex)
func (msign *OldMintSign) Sign(dMinter *esign.ERC712, pk string) ([]byte, error) {
	signedRaw, err := dMinter.Sign(pk, map[string]interface{}{
		"iot":    msign.Iot,
		"amount": msign.Amount,
		"nonce":  msign.Nonce,
	})
	if nil != err {
		return nil, err
	}

	msign.R = hexutil.Encode(signedRaw[:32])
	msign.S = hexutil.Encode(signedRaw[32:64])
	msign.V = hexutil.Encode(signedRaw[64:])

	return signedRaw, nil
}

func (msign *OldMintSign) Verify(dMinter *esign.ERC712) error {
	var data = map[string]interface{}{
		"iot":    msign.Iot,
		"amount": msign.Amount,
		"nonce":  msign.Nonce,
	}

	var signed, err = hexutil.Decode(
		esign.HexConcat(msign.R, msign.S, msign.V),
	)

	if nil != err {
		return dmodels.NewError(ecodes.IOTInvalidMintSign, "Invalid mint sign: "+err.Error())
	}

	err = dMinter.Verify(msign.Iot, signed, data)
	if nil != err {
		return dmodels.NewError(ecodes.IOTInvalidMintSign, "Invalid mint sign: "+err.Error())
	}
	return nil
}

func ConvertMintSign(oms *OldMintSign) (*models.MintSign, error) {
	raw, err := hexutil.Decode(esign.HexConcat(oms.R, oms.S, oms.V))
	if nil != err {
		return nil, err
	}
	var rs = &models.MintSign{
		Id:        oms.ID,
		IotId:     oms.IotId,
		Nonce:     oms.Nonce,
		Amount:    oms.Amount,
		Iot:       oms.Iot,
		Signed:    base64.StdEncoding.EncodeToString(raw),
		CreatedAt: oms.CreatedAt,
		UpdatedAt: oms.UpdatedAt,
	}
	return rs, nil
}

type OldMinted struct {
	ID        string    `json:"id,omitempty" `
	IotId     int64     `json:"iotId,omitempty" gorm:"index:minted_idx_ca_iot,priority:2"`
	Carbon    int64     `json:"carbon,omitempty" `
	CreatedAt time.Time `json:"createdAt,omitempty" gorm:"index:minted_idx_ca_iot,priority:1"`
}

func (*OldMinted) TableName() string { return OldTableNameMinted }

func ConvertMinted(ominted *OldMinted) (*models.Minted, error) {
	return &models.Minted{
		Id:        ominted.ID,
		IotId:     ominted.IotId,
		Carbon:    ominted.Carbon,
		CreatedAt: ominted.CreatedAt,
	}, nil
}
