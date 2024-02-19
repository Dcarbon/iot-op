package repo

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/domain/rss"
	"github.com/Dcarbon/iot-op/internal/models"
)

var stateTest *StateImpl
var pkey = "0123456789012345678901234567890123456789012345678901234567880000"

func TestStateUpdate(t *testing.T) {
	var data = &models.StateExtract{
		State: models.StateIdle,
		Sensors: []*models.Sensor{
			{
				State:  models.StateActived,
				Metric: map[string]float64{"value": 1},
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

	raw, err := json.Marshal(data)
	utils.PanicError("", err)

	signedRaw, err := esign.SignPersonal(pkey, raw)
	utils.PanicError("", err)
	log.Println("Signed raw: ", string(signedRaw))

	var req = &domain.RStateUpdate{
		Signature: dmodels.Signature{
			Signer: dmodels.EthAddress(addr),
			Data:   base64.StdEncoding.EncodeToString(raw),
			Signed: base64.StdEncoding.EncodeToString(signedRaw),
		},
	}
	err = getStateTest().Update(req)
	utils.PanicError("", err)
}

func TestStateGet(t *testing.T) {
	data, err := getStateTest().Get(&domain.RStateGet{
		IotId: 0,
	})
	utils.PanicError("", err)
	utils.Dump("", data)
}

func getStateTest() *StateImpl {
	if stateTest != nil {
		return stateTest
	}

	var err error
	stateTest, err = NewStateImpl(
		rss.GetRedis(),
		svc.NewMockIotClient(svc.DefaultMockIot...),
	)
	utils.PanicError("", err)
	return stateTest
}
