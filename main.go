package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Dcarbon/arch-proto/pb"
	"github.com/Dcarbon/go-shared/gutils"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/iot-op/internal/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"google.golang.org/grpc"
)

var config = gutils.Config{
	Port:   utils.IntEnv("PORT", 4003),
	DbUrl:  utils.StringEnv("DB_URL", "postgres://admin:hellosecret@13.228.11.143/iot_op"),
	Name:   "IotOperator", // Sensor service
	JwtKey: utils.StringEnv("JWT", ""),
	Options: map[string]string{
		gutils.ISVIotInfo: utils.StringEnv(gutils.ISVIotInfo, "localhost:4002"),
		"AMQP_URL":        utils.StringEnv("AMQP_URL", "amqp://rbuser:hellosecret@localhost"),
		"REDIS_URL":       utils.StringEnv("REDIS_URL", "redis://localhost:6379"),
		"CHAIN_ID":        utils.StringEnv("CHAIN_ID", "1337"),
		"CARBON_NAME":     utils.StringEnv("CARBON_NAME", "Carbon"),
		"CARBON_VERSION":  utils.StringEnv("CARBON_VERSION", "1"),
		"CARBON_ADDRESS": utils.StringEnv(
			"CARBON_ADDRESS",
			"0x9C399C33a393334D28e8bA4FFF45296f50F82d1f",
		),
	},
	AuthConfig: map[string]*gutils.ARConfig{
		"/pb.IotOpService/CreateMint": {
			Require:    false,
			Permission: "iot-op-create-mint",
			PermDesc:   "",
		},
		"/pb.IotOpService/GetMintSigns": {
			Require:    false,
			Permission: "iot-op-get-mint-sign",
			PermDesc:   "",
		},
		"/pb.IotOpService/GetMintSignLatest": {
			Require:    false,
			Permission: "iot-op-get-mint-sign-latest",
			PermDesc:   "",
		},
		"/pb.IotOpService/GetMinted": {
			Require:    false,
			Permission: "iot-op-",
			PermDesc:   "",
		},
		"/pb.IotOpService/CreateAVM": {
			Require:    false,
			Permission: "iot-op-",
			PermDesc:   "",
		},
		"/pb.IotOpService/GetListAVM": {
			Require:    false,
			Permission: "iot-op-",
			PermDesc:   "",
		},
		"/pb.IotOpService/UpdateState": {
			Require:    false,
			Permission: "iot-op-",
			PermDesc:   "",
		},
		"/pb.IotOpService/GetState": {
			Require:    false,
			Permission: "iot-op-",
			PermDesc:   "",
		},

		"/pb.IotOpService/GetSeparator": {
			Require:    false,
			Permission: "",
			PermDesc:   "",
		},
		"/pb.IotOpService/SetVersion": {
			Require:    false,
			Permission: "",
			PermDesc:   "",
		},
		"/pb.IotOpService/GetVersion": {
			Require:    false,
			Permission: "",
			PermDesc:   "",
		},
		"/pb.IotOpService/IsIotActivated": {
			Require:    false,
			Permission: "",
			PermDesc:   "",
		},
		"/pb.IotOpService/Offset": {
			Require:    false,
			Permission: "",
			PermDesc:   "",
		},
	},
}

func main() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	utils.PanicError(config.Name+" open port", err)

	logger := gutils.NewLogInterceptor()
	auth, err := gutils.NewAuthInterceptor(
		// config.GetIAM(),
		config.JwtKey,
		config.AuthConfig,
	)
	utils.PanicError("Authen init", err)

	handler, err := service.NewService(&config)
	utils.PanicError(config.Name+" init", err)

	var sv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			gutils.UnaryPreventPanic,
			logger.Intercept,
			validator.UnaryServerInterceptor(),
			auth.Intercept,
		),
	)

	pb.RegisterIotOpServiceServer(sv, handler)
	log.Println(config.Name+" listen and serve at ", config.Port)
	err = sv.Serve(listen)
	if nil != err {
		log.Fatal(config.Name+" listen and serve error: ", err)
	}
}
