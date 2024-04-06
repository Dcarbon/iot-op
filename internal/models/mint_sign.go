package models

import (
	"encoding/base64"
	"encoding/binary"
	"log"
	"strconv"
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/ecodes"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	TableNameMintSign = "mint_sign"
	TableNameMinted   = "minted"
)

type MintSign struct {
	Id        int64     `json:"id"        gorm:"primary_key"` //
	IotId     int64     `json:"iotId"     `                   // IoT id
	Nonce     int64     `json:"nonce"     gorm:"index"`       //
	Amount    string    `json:"amount"    `                   // Hex
	Iot       string    `json:"iot"       gorm:"index"`       // IoT Address
	Signed    string    `json:"signed"    `                   // Base 64
	CreatedAt time.Time `json:"createdAt" `                   //
	UpdatedAt time.Time `json:"updatedAt" `                   //
	// R         string    `json:"r" `         //
	// S         string    `json:"s" `         //
	// V         string    `json:"v" `         //
}

func (*MintSign) TableName() string { return TableNameMintSign }

// Only for test
// pk: private key (hex)
func (msign *MintSign) Sign(dMinter *esign.ERC712, pk string) ([]byte, error) {
	signedRaw, err := dMinter.Sign(pk, map[string]interface{}{
		"iot":    msign.Iot,
		"amount": msign.Amount,
		"nonce":  msign.Nonce,
	})
	if nil != err {
		return nil, err
	}
	msign.Signed = base64.StdEncoding.EncodeToString(signedRaw)
	// msign.R = hexutil.Encode(signedRaw[:32])
	// msign.S = hexutil.Encode(signedRaw[32:64])
	// msign.V = hexutil.Encode(signedRaw[64:])

	return signedRaw, nil
}

func (msign *MintSign) Verify(dMinter *esign.ERC712) error {
	var data = map[string]interface{}{
		"iot":    msign.Iot,
		"amount": msign.Amount,
		"nonce":  msign.Nonce,
	}

	var signed, err = base64.StdEncoding.DecodeString(msign.Signed)
	if nil != err {
		return dmodels.NewError(ecodes.IOTInvalidMintSign, "Invalid mint sign: "+err.Error())
	}
	log.Println("Signed on verify: ", hexutil.Encode(signed))
	err = dMinter.Verify(msign.Iot, signed, data)
	if nil != err {
		return dmodels.NewError(ecodes.IOTInvalidMintSign, "Invalid mint sign: "+err.Error())
	}
	return nil
}

func (msign *MintSign) VerifySolana() error {
	var signed, err = base64.StdEncoding.DecodeString(msign.Signed)
	if nil != err {
		return dmodels.NewError(
			ecodes.IOTInvalidMintSign,
			"Invalid signature. It must be base64: "+err.Error(),
		)
	}

	buff, err := msign.GenerateSignData()
	if nil != err {
		return err
	}

	err = esign.VerifyPersonalSign(msign.Iot, crypto.Keccak256(buff), signed)
	if nil != err {
		return dmodels.NewError(ecodes.IOTInvalidMintSign, "Invalid mint sign: "+err.Error())
	}

	return nil
}

func (msign *MintSign) GenerateSignData() ([]byte, error) {
	addr, err := hexutil.Decode(msign.Iot)
	if nil != err {
		return nil, err
	}

	amount, err := strconv.ParseUint(msign.Amount[2:], 16, 64)
	if nil != err {
		return nil, err
	}

	var buff = []byte{}

	buff = append(buff, addr...)
	buff = binary.BigEndian.AppendUint64(buff, uint64(amount))
	buff = binary.BigEndian.AppendUint64(buff, uint64(msign.Nonce))

	// buff = append(buff, addr...)
	// buff = append(buff, []byte(strconv.FormatUint(amount, 10))...)
	// buff = append(buff, []byte(strconv.FormatUint(nonce, 10))...)

	return buff, nil
}

type Minted struct {
	Id        string    `json:"id,omitempty" `
	IotId     int64     `json:"iotId,omitempty" gorm:"index:minted_idx_ca_iot,priority:2"`
	Carbon    int64     `json:"carbon,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" gorm:"index:minted_idx_ca_iot,priority:1"`
}

func (*Minted) TableName() string { return TableNameMinted }
