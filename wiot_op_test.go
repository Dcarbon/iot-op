package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/Dcarbon/arch-proto/pb"
	"github.com/Dcarbon/go-shared/gutils"
	"github.com/Dcarbon/go-shared/libs/decimal"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/models"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var reqCtx = context.TODO()
var host = utils.StringEnv("IOT_OP_HOST", "localhost:4003")

var iotOpClient pb.IotOpServiceClient
var pkey = "0123456789012345678901234567890123456789012345678901234567880000"

func init() {
	go main()
	time.Sleep(3 * time.Second)

	cc, err := gutils.GetCCTimeout(host, 5*time.Second)
	utils.PanicError("Connect to service timeout", err)

	iotOpClient = pb.NewIotOpServiceClient(cc)
}

func TestAVMCreate(t *testing.T) {
	var x = &domain.AVMSignExtract{
		Volume: decimal.NewFromFloat(100.1),
		From:   time.Now().Unix() - 30*86400,
		To:     time.Now().Unix() - 29*86400,
	}

	addr, err := esign.GetAddress(pkey)
	utils.PanicError("", err)

	raw, err := json.Marshal(x)
	utils.PanicError("", err)

	signedRaw, err := esign.SignPersonal(pkey, raw)
	utils.PanicError("", err)
	log.Println("Signed raw: ", string(signedRaw))

	var req = &pb.RAVMCreate{
		Signer: addr,
		Data:   base64.StdEncoding.EncodeToString(raw),
		Signed: base64.StdEncoding.EncodeToString(signedRaw),
	}

	data, err := iotOpClient.CreateAVM(reqCtx, req)
	utils.PanicError("", err)
	utils.Dump("Create avm", data)
}

func TestAVMGetList(t *testing.T) {
	var req = &pb.RAVMGetList{
		Full:     false,
		Skip:     0,
		Limit:    1,
		IotId:    2,
		From:     time.Now().Unix() - 30*86400,
		To:       time.Now().Unix(),
		Sort:     pb.Sort_SortASC,
		Interval: pb.DInterval_DI_Day,
	}
	var data, err = iotOpClient.GetListAVM(reqCtx, req)
	utils.PanicError("", err)
	utils.Dump("", data)
}

func TestMintSignCreate(t *testing.T) {
	separator, err := iotOpClient.GetSeparator(reqCtx, &pb.Empty{})
	utils.PanicError("", err)
	utils.Dump("separator: ", separator)

	var typedDomain = &esign.TypedDataDomain{
		Name:              "Carbon",
		ChainId:           separator.ChainId,
		Version:           separator.Version,
		VerifyingContract: separator.VerifyingContract,
	}

	utils.Dump("TypeDomain Client ", typedDomain)

	var dMinter = esign.MustNewERC712(
		typedDomain,
		esign.MustNewTypedDataField(
			"Mint",
			esign.TypedDataStruct,
			esign.MustNewTypedDataField("iot", esign.TypedDataAddress),
			esign.MustNewTypedDataField("amount", esign.TypedDataUint256),
			esign.MustNewTypedDataField("nonce", esign.TypedDataUint256),
		),
	)
	var amount uint64 = 2 * 1e9
	var nonce int64 = 1
	var amountHex = hexutil.EncodeUint64(uint64(amount))

	addr, err := esign.GetAddress(pkey)
	utils.PanicError("", err)
	log.Println("Address: ", addr)

	var msign = map[string]interface{}{
		"iot":    addr,
		"amount": amountHex,
		"nonce":  nonce,
	}

	signed, err := dMinter.Sign(pkey, msign)
	utils.PanicError("Minter signed", err)

	log.Println("Signed on Test: ", hexutil.Encode(signed))

	req := &pb.RIotMint{
		Nonce:  nonce,
		Amount: amountHex,
		Iot:    addr,
		Signed: base64.StdEncoding.EncodeToString(signed),
	}

	_, err = iotOpClient.CreateMint(reqCtx, req)
	utils.PanicError("", err)
}

func TestMintSignGetList(t *testing.T) {
	var req = &pb.RIotGetMintSigns{
		From:  time.Now().Unix() - 30*86400,
		To:    time.Now().Unix(),
		IotId: 291,
		Sort:  pb.Sort_SortDesc,
		Limit: 4,
	}
	var data, err = iotOpClient.GetMintSigns(reqCtx, req)
	utils.PanicError("", err)
	utils.Dump("", data)
}

func TestMintedGetList(t *testing.T) {
	var req = &pb.RIotGetMinted{
		IotId:    291,
		From:     time.Now().Unix() - 30*86400,
		To:       time.Now().Unix(),
		Interval: int32(pb.DInterval_DI_Month),
	}
	var data, err = iotOpClient.GetMinted(reqCtx, req)
	utils.PanicError("", err)
	utils.Dump("", data)
}

func TestUpdateState(t *testing.T) {
	var state = &models.StateExtract{
		State: models.StateActived,
		Sensors: []*models.Sensor{
			{
				State:  models.StateActived,
				Metric: map[string]float64{"temperature": 80},
				// Metric: dmodels.AllMetric{
				// 	DefaultMetric: dmodels.DefaultMetric{
				// 		Val: 1,
				// 	},
				// },
			},
		},
	}
	addr, err := esign.GetAddress(pkey)
	utils.PanicError("", err)

	raw, err := json.Marshal(state)
	utils.PanicError("", err)

	signedRaw, err := esign.SignPersonal(pkey, raw)
	utils.PanicError("", err)

	var req = &pb.RIotUpdateState{
		Signer: addr,
		Data:   base64.StdEncoding.EncodeToString(raw),
		Signed: base64.StdEncoding.EncodeToString(signedRaw),
	}

	_, err = iotOpClient.UpdateState(reqCtx, req)
	utils.PanicError("", err)
}

func TestGetState(t *testing.T) {
	var req = &pb.IdInt64{
		Id: 2,
	}
	var data, err = iotOpClient.GetState(reqCtx, req)
	utils.PanicError("", err)
	utils.Dump("Dump state", data)
}

func TestSetVersion(t *testing.T) {
	_, err := iotOpClient.SetVersion(reqCtx, &pb.RIotSetVersion{
		IotType: 1,
		Version: "0.1.0",
	})
	utils.PanicError("", err)

	v, err := iotOpClient.GetVersion(reqCtx, &pb.RIotGetVersion{
		IotType: 1,
	})
	utils.PanicError("", err)
	utils.Dump("Current version: ", v)
}

func TestGetVersion(t *testing.T) {

}
