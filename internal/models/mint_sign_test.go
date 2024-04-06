package models

import (
	"encoding/binary"
	"log"
	"testing"

	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestGenerateSign(t *testing.T) {
	var msign = &MintSign{
		Id:     1,
		IotId:  1,
		Iot:    "0x7cFdF09A834fd1143238D594e9B6179861734ba1",
		Nonce:  123,
		Amount: "0x64",
	}
	var buff, err = msign.GenerateSignData()
	utils.PanicError("err", err)
	log.Println("Msg: ", hexutil.Encode(buff))

	var hash = crypto.Keccak256(buff)
	log.Println("Hash: ", hexutil.Encode(hash))
}

func TestAA(t *testing.T) {
	// var buff = []byte{}
	var buff, err = hexutil.Decode("0x7cFdF09A834fd1143238D594e9B6179861734ba1")
	utils.PanicError("", err)

	var amount = 100
	var nonce = 123
	buff = binary.BigEndian.AppendUint64(buff, uint64(amount))
	buff = binary.BigEndian.AppendUint64(buff, uint64(nonce))

	var hash = crypto.Keccak256(buff)
	log.Println("Hash size: ", len(hash))
	log.Println("Hash: ", hexutil.Encode(hash))

}

func TestSign(t *testing.T) {
	var msign = &MintSign{
		Id:     1,
		IotId:  1,
		Iot:    "0x7cFdF09A834fd1143238D594e9B6179861734ba1",
		Nonce:  123,
		Amount: "0x64",
		Signed: "sjdHqT0Zr2p0aaGWX9YYQkwzDdvHQxhrma4z/I0N38cYg0aTnCu0vMXu+3DhT/fMaMmZ6Yjn9y99WI8BfNz+5xw=",
	}

	err := msign.VerifySolana()
	utils.PanicError("Verify error", err)

}

func TestSign2(t *testing.T) {
	var msign = &MintSign{
		Id:     1,
		IotId:  1,
		Iot:    "0x6d8df15be099dea003fb3944a152abb9975c6fda",
		Nonce:  1,
		Amount: "0x07f8930131",
		Signed: "RM4zQVFAHB7CnQMenxYxNhFpH9H88HJFqpCorm3lK7JoApBitTC4j5/s8txhXUEBCeEQ37iZUJvHTByGWPqcBBw=",
	}

	err := msign.VerifySolana()
	utils.PanicError("Verify error", err)

}
