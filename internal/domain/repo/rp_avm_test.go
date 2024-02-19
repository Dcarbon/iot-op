package repo

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/libs/decimal"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/domain/rss"
)

var avmTest *AVMImpl

func TestAVMCreate(t *testing.T) {
	var x = &domain.AVMSignExtract{
		Volume: decimal.NewFromFloat(107.1),
		From:   time.Now().Unix() - 1*86400 + 5*60,
		To:     time.Now().Unix() - 1*86400 + 6*60,
	}

	addr, err := esign.GetAddress(pkey)
	utils.PanicError("", err)

	raw, err := json.Marshal(x)
	utils.PanicError("", err)

	signedRaw, err := esign.SignPersonal(pkey, raw)
	utils.PanicError("", err)

	var req = &domain.RAVMCreate{
		Signature: dmodels.Signature{
			Signer: dmodels.EthAddress(addr),
			Data:   base64.StdEncoding.EncodeToString(raw),
			Signed: base64.StdEncoding.EncodeToString(signedRaw),
		},
	}
	data, err := getAVMTest().Create(req)
	utils.PanicError("", err)
	utils.Dump("", data)
}

func TestAVMGetList(t *testing.T) {
	var req = &domain.RAVMGetList{
		Full:     true,
		Skip:     0,
		Limit:    3,
		IotId:    2,
		From:     time.Now().Unix() - 60*86400,
		To:       time.Now().Unix(),
		Sort:     dmodels.SortASC,
		Interval: dmodels.DINone,
	}
	data, err := getAVMTest().GetList(req)
	utils.PanicError("", err)
	utils.Dump("", data)
}

func getAVMTest() *AVMImpl {
	if avmTest != nil {
		return avmTest
	}

	var err error
	avmTest, err = NewAVMImpl(rss.GetDB(),
		svc.NewMockIotClient(svc.DefaultMockIot...),
	)
	utils.PanicError("", err)
	return avmTest
}
