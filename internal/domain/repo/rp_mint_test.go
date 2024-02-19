package repo

import (
	"encoding/base64"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/domain/rss"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var minterTest *MintImpl

// var testDomainMinter = esign.MustNewERC712(
// 	&esign.TypedDataDomain{
// 		Name:              "CARBON",
// 		Version:           "1",
// 		ChainId:           1337,
// 		VerifyingContract: "0x7BDDCb9699a3823b8B27158BEBaBDE6431152a85",
// 	},
// 	esign.MustNewTypedDataField(
// 		"Mint",
// 		esign.TypedDataStruct,
// 		esign.MustNewTypedDataField("iot", esign.TypedDataAddress),
// 		esign.MustNewTypedDataField("amount", "uint256"),
// 		esign.MustNewTypedDataField("nonce", "uint256"),
// 	),
// )

var typedDomain = &esign.TypedDataDomain{
	Name:              "Carbon",
	ChainId:           utils.Int64Env("CHAIN_ID", 1337),
	Version:           utils.StringEnv("CARBON_VERSION", "1"),
	VerifyingContract: utils.StringEnv("CARBON_ADDRESS", "0x9C399C33a393334D28e8bA4FFF45296f50F82d1f"),
}

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

func TestMinterMint(t *testing.T) {
	var minterTest = getMinterTest()
	var amount uint64 = 2 * 1e9

	addr, err := esign.GetAddress(pkey)
	utils.PanicError("", err)
	log.Println("Address: ", addr)

	var msign = map[string]interface{}{
		"iot":    strings.ToLower(addr),
		"amount": hexutil.EncodeUint64(uint64(amount)),
		"nonce":  1,
	}
	log.Println("Type domain: ", dMinter)

	signed, err := minterTest.dMinter.Sign(pkey, msign)
	utils.PanicError("Minter signed", err)

	log.Println("Signed on Test: ", hexutil.Encode(signed))

	// signed, err := sign.Sign(dMinter, pkey)
	// utils.PanicError("TestCreateMint", err)

	err = minterTest.dMinter.Verify(addr, signed, msign)
	utils.PanicError("", err)

	var req = &domain.RMinterMint{
		Nonce:  1,
		Amount: hexutil.EncodeUint64(uint64(amount)),
		Iot:    addr,
		Signed: base64.StdEncoding.EncodeToString(signed),
	}

	err = minterTest.Mint(req)
	utils.PanicError("", err)
	// utils.Dump("", data)
}

func TestMinterGetSigned(t *testing.T) {}

func TestMinterGetMinted(t *testing.T) {
	var req = &domain.RMinterGetMinted{
		Skip:     0,
		Limit:    2,
		Sort:     dmodels.SortASC,
		Interval: dmodels.DIMonth,
		IotId:    291,
		From:     time.Now().Unix() - 60*86400,
	}
	data, err := getMinterTest().GetMinted(req)
	utils.PanicError("", err)
	utils.Dump("", data)

}

func getMinterTest() *MintImpl {
	if minterTest != nil {
		return minterTest
	}
	var err error
	minterTest, err = NewMinterImpl(
		rss.GetDB(),
		dMinter,
		svc.NewMockIotClient(svc.DefaultMockIot...),
	)
	utils.PanicError("", err)
	return minterTest
}
