package service

import (
	"context"
	"fmt"

	"github.com/Dcarbon/arch-proto/pb"
	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/gutils"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/domain/repo"
	"github.com/Dcarbon/iot-op/internal/domain/rss"
	"github.com/Dcarbon/iot-op/internal/models"
	"google.golang.org/grpc/peer"
)

type Service struct {
	pb.UnimplementedIotOpServiceServer
	*gutils.GService
	iot      svc.IIotInfo
	iavm     domain.IAVM
	iminter  domain.IMinter
	istate   domain.IState
	iversion domain.IVersion
}

func NewService(config *gutils.Config) (*Service, error) {
	rss.SetUrl(config.GetDBUrl(), config.GetRedisUrl())

	var typedDomain = &esign.TypedDataDomain{
		Name:              config.GetOption("CARBON_NAME"),
		ChainId:           config.GetOptInt("CHAIN_ID"),
		Version:           config.GetOption("CARBON_VERSION"),
		VerifyingContract: config.GetOption("CARBON_ADDRESS"),
	}

	dMinter, err := esign.NewERC712(
		typedDomain,
		esign.MustNewTypedDataField(
			"Mint",
			esign.TypedDataStruct,
			esign.MustNewTypedDataField("iot", esign.TypedDataAddress),
			esign.MustNewTypedDataField("amount", esign.TypedDataUint256),
			esign.MustNewTypedDataField("nonce", esign.TypedDataUint256),
		),
	)
	if nil != err {
		return nil, err
	}
	utils.Dump("TypeDomainService: ", typedDomain)

	iot, err := svc.NewIotService(config.GetIotHost())
	if nil != err {
		return nil, err
	}

	iavm, err := repo.NewAVMImpl(rss.GetDB(), iot)
	if nil != err {
		return nil, err
	}

	iminter, err := repo.NewMinterImpl(rss.GetDB(), dMinter, iot)
	if nil != err {
		return nil, err
	}

	istate, err := repo.NewStateImpl(rss.GetRedis(), iot)
	if nil != err {
		return nil, err
	}

	sv := &Service{
		iot:     iot,
		iavm:    iavm,
		iminter: iminter,
		istate:  istate,
	}

	sv.iversion, err = repo.NewVersionImpl(sv.initVersion(), sv.initDownload())
	if nil != err {
		return nil, err
	}

	return sv, nil
}

func (sv *Service) CreateAVM(ctx context.Context, req *pb.RAVMCreate,
) (*pb.AVM, error) {
	data, err := sv.iavm.Create(&domain.RAVMCreate{
		Signature: dmodels.Signature{
			Signer: dmodels.EthAddress(req.Signer),
			Signed: req.Signed,
			Data:   req.Data,
		},
	})
	if nil != err {
		return nil, err
	}

	return convertAVM(data), nil
}

func (sv *Service) GetListAVM(ctx context.Context, req *pb.RAVMGetList,
) (*pb.AVMs, error) {
	data, err := sv.iavm.GetList(&domain.RAVMGetList{
		Full:     req.Full,
		Skip:     int(req.Skip),
		Limit:    int(req.Limit),
		IotId:    req.IotId,
		From:     req.From,
		To:       req.To,
		Sort:     dmodels.Sort(req.Sort),
		Interval: dmodels.DInterval(req.Interval),
	})
	if nil != err {
		return nil, err
	}

	return &pb.AVMs{
		Data: convertArr[models.AVM, pb.AVM](data, convertAVM),
	}, nil
}

func (sv *Service) CreateMint(ctx context.Context, req *pb.RIotMint,
) (*pb.Empty, error) {
	err := sv.iminter.Mint(&domain.RMinterMint{
		Nonce:  req.Nonce,
		Amount: req.Amount,
		Iot:    req.Iot,
		Signed: req.Signed,
	})
	if nil != err {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (sv *Service) GetMintSigns(ctx context.Context, req *pb.RIotGetMintSigns,
) (*pb.MintedSigns, error) {
	signeds, err := sv.iminter.GetSigns(&domain.RMinterGetSigns{
		From:  req.From,
		To:    req.To,
		IotId: req.IotId,
		Sort:  dmodels.Sort(req.Sort),
		Limit: int(req.Limit),
	})
	if nil != err {
		return nil, err
	}

	return &pb.MintedSigns{
		Total: 0,
		Data:  convertArr(signeds, convertMintedSign),
	}, nil
}

func (sv *Service) GetMint(ctx context.Context, req *pb.RIotGetMintSigns,
) (*pb.MintedSigns, error) {
	signeds, err := sv.iminter.GetSign(&domain.RMinterGetSigns{
		From:  req.From,
		To:    req.To,
		IotId: req.IotId,
		Sort:  dmodels.Sort(req.Sort),
		Limit: int(req.Limit),
	})
	if nil != err {
		return nil, err
	}

	return &pb.MintedSigns{
		Total: 0,
		Data:  convertArr(signeds, convertMintedSign),
	}, nil
}

func (sv *Service) GetMintSignLatest(ctx context.Context, req *pb.RIotGetMintSignLatest,
) (*pb.MintedSign, error) {
	signed, err := sv.iminter.GetSignLatest(&domain.RMinterGetSignLatest{
		IotId: req.IotId,
	})
	if nil != err {
		return nil, err
	}
	if signed.Id == 0 {
		return nil, gutils.ErrNotFound("Has no sign")
	}

	return convertMintedSign(signed), nil
}

func (sv *Service) IsIotActivated(ctx context.Context, req *pb.RIotIsActivated,
) (*pb.Bool, error) {
	activated, err := sv.iminter.IsIotActivated(&domain.RIsIotActivated{
		From:  req.From,
		To:    req.To,
		IotId: req.IotId,
	})
	if nil != err {
		return nil, err
	}
	return &pb.Bool{Data: activated}, nil
}

func (sv *Service) GetMinted(ctx context.Context, req *pb.RIotGetMinted,
) (*pb.IotMinteds, error) {
	data, err := sv.iminter.GetMinted(&domain.RMinterGetMinted{
		From:     req.From,
		To:       req.To,
		IotId:    req.IotId,
		Interval: dmodels.DInterval(req.Interval),
		Limit:    req.Limit,
	})
	if nil != err {
		return nil, err
	}

	var rs = &pb.IotMinteds{
		Data: convertArr(data, convertMinted),
	}

	return rs, nil
}

func (sv *Service) UpdateState(ctx context.Context, req *pb.RIotUpdateState,
) (*pb.Empty, error) {
	err := sv.istate.Update(&domain.RStateUpdate{
		Signature: dmodels.Signature{
			Signer: dmodels.EthAddress(req.Signer),
			Signed: req.Signed,
			Data:   req.Data,
		},
	})

	if nil != err {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (sv *Service) GetState(ctx context.Context, req *pb.IdInt64,
) (*pb.IotState, error) {
	state, err := sv.istate.Get(&domain.RStateGet{
		IotId: req.Id,
	})
	if nil != err {
		return nil, err
	}

	var rs = &pb.IotState{
		State:     int32(state.State),
		Sensors:   make([]*pb.SensorState, len(state.Sensors)),
		CreatedAt: state.CreatedAt,
		Info:      state.Info,
	}

	for i, it := range state.Sensors {
		rs.Sensors[i] = &pb.SensorState{
			State:  int32(it.State),
			Metric: it.Metric,
			Type:   int32(it.Type),
		}
	}
	return rs, nil
}

func (sv *Service) GetSeparator(ctx context.Context, req *pb.Empty,
) (*pb.Separator, error) {
	var td2, err = sv.iminter.GetSeparator()
	if nil != err {
		return nil, err
	}

	var rs = &pb.Separator{
		Name:              td2.Name,
		Version:           td2.Version,
		ChainId:           td2.ChainId,
		VerifyingContract: td2.VerifyingContract,
	}
	return rs, nil
}

func (sv *Service) SetVersion(ctx context.Context, req *pb.RIotSetVersion,
) (*pb.Empty, error) {
	err := sv.iversion.SetVersion(&domain.RVersionSet{
		IotType: req.IotType,
		Version: req.Version,
	})
	if nil != err {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (sv *Service) GetVersion(ctx context.Context, req *pb.RIotGetVersion,
) (*pb.RsIotVersion, error) {
	version, path, err := sv.iversion.GetVersion(&domain.RVersionGet{
		IotType: req.IotType,
	})
	p, _ := peer.FromContext(ctx)
	fmt.Println("IP Called: " + p.Addr.String())

	if nil != err {
		return nil, err
	}

	return &pb.RsIotVersion{
		Version: version,
		Path:    path,
	}, nil
}

func (sv *Service) Offset(ctx context.Context, req *pb.RIotOffset) (*pb.RsIotOffset, error) {

	res, err := sv.iminter.MintedOffset(domain.RIotOffset{})
	if nil != err {
		return nil, err
	}
	return &pb.RsIotOffset{Amount: res.Amount}, nil
}

func (sv *Service) initVersion() map[int32]string {
	var rs = map[int32]string{
		int32(dmodels.IotTypeWindPower): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d", dmodels.IotTypeWindPower),
			"0.0.1",
		),
		int32(dmodels.IotTypeSolarPower): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d", dmodels.IotTypeSolarPower),
			"0.0.1",
		),
		int32(dmodels.IotTypeBurnMethane): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d", dmodels.IotTypeBurnMethane),
			"0.0.2",
		),
		int32(dmodels.IotTypeBurnBiomass): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d", dmodels.IotTypeBurnBiomass),
			"0.0.6",
		),
		int32(dmodels.IotTypeFertilizer): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d", dmodels.IotTypeFertilizer),
			"0.0.1",
		),
		int32(dmodels.IotTypeBurnTrash): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d", dmodels.IotTypeBurnTrash),
			"0.0.1",
		),
	}
	return rs
}

func (sv *Service) initDownload() map[int32]string {
	var rs = map[int32]string{
		int32(dmodels.IotTypeWindPower): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d_PATH", dmodels.IotTypeWindPower),
			fmt.Sprintf("%s/static/iots/ota/%d/0.0.1",
				utils.StringEnv(gutils.EXTERNAL_HOST, "http://localhost:4000"),
				dmodels.IotTypeWindPower,
			),
		),

		int32(dmodels.IotTypeSolarPower): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d_PATH", dmodels.IotTypeSolarPower),
			fmt.Sprintf("%s/static/iots/ota/%d/0.0.1",
				utils.StringEnv(gutils.EXTERNAL_HOST, "http://localhost:4000"),
				dmodels.IotTypeSolarPower,
			),
		),
		int32(dmodels.IotTypeBurnMethane): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d_PATH", dmodels.IotTypeBurnMethane),
			fmt.Sprintf("%s/static/iots/ota/%d/0.0.1",
				utils.StringEnv(gutils.EXTERNAL_HOST, "http://localhost:4000"),
				dmodels.IotTypeBurnMethane,
			),
		),
		int32(dmodels.IotTypeBurnBiomass): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d_PATH", dmodels.IotTypeBurnBiomass),
			fmt.Sprintf("%s/static/iots/ota/%d/0.0.5",
				utils.StringEnv(gutils.EXTERNAL_HOST, "http://localhost:4000"),
				dmodels.IotTypeBurnBiomass,
			),
		),
		int32(dmodels.IotTypeFertilizer): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d_PATH", dmodels.IotTypeFertilizer),
			fmt.Sprintf("%s/static/iots/ota/%d/0.0.1",
				gutils.EXTERNAL_HOST,
				dmodels.IotTypeFertilizer,
			),
		),
		int32(dmodels.IotTypeBurnTrash): utils.StringEnv(
			fmt.Sprintf("VERSION_IOT_%d_PATH", dmodels.IotTypeBurnTrash),
			fmt.Sprintf("%s/static/iots/ota/%d/0.0.1",
				utils.StringEnv(gutils.EXTERNAL_HOST, "http://localhost:4000"),
				dmodels.IotTypeBurnTrash,
			),
		),
	}
	return rs
}
