package service

import (
	"github.com/Dcarbon/arch-proto/pb"
	"github.com/Dcarbon/iot-op/internal/models"
)

func convertAVM(in *models.AVM) *pb.AVM {
	var rs = &pb.AVM{
		Id:        in.Id,
		Signed:    in.Signed,
		Data:      in.Data,
		Value:     in.Volume.StringFixed(2),
		CreatedAt: in.CreatedAt.UnixMilli(),
	}
	return rs
}

func convertMintedSign(in *models.MintSign,
) *pb.MintedSign {
	var rs = &pb.MintedSign{
		Id:        in.Id,
		IotId:     in.IotId,
		Nonce:     in.Nonce,
		Amount:    in.Amount,
		Signed:    in.Signed,
		CreatedAt: in.CreatedAt.UnixMilli(),
		UpdatedAt: in.CreatedAt.UnixMilli(),
	}
	return rs
}

func convertMinted(in *models.Minted) *pb.IotMinted {
	var rs = &pb.IotMinted{
		Id:        in.Id,
		IotId:     in.IotId,
		Carbon:    in.Carbon,
		CreatedAt: in.CreatedAt.UnixMilli(),
	}
	return rs
}

func convertArr[T any, T2 any](arr []*T, fn func(*T) *T2) []*T2 {
	var rs = make([]*T2, len(arr))
	for i, it := range arr {
		rs[i] = fn(it)
	}
	return rs
}
