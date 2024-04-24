package repo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dcarbon/go-shared/gutils"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/models"
	"github.com/go-redis/redis/v8"
)

var stateDuration = 500 * time.Second

type StateImpl struct {
	rClient *redis.Client
	iot     svc.IIotInfo
}

func NewStateImpl(rClient *redis.Client, iiot svc.IIotInfo) (*StateImpl, error) {
	var impl = &StateImpl{
		rClient: rClient,
		iot:     iiot,
	}
	return impl, nil
}

func (impl *StateImpl) Update(req *domain.RStateUpdate,
) error {
	err := req.Verify()
	if nil != err {
		return err
	}

	stateRaw, err := base64.StdEncoding.DecodeString(req.Data)
	if nil != err {
		return gutils.ErrBadRequest("Invalid state format " + err.Error())
	}
	// log.Println("state raw: ", string(stateRaw))

	var state = &models.StateExtract{}
	err = json.Unmarshal(stateRaw, state)
	if nil != err {
		return err
	}
	state.CreatedAt = time.Now().Unix()

	iot, err := impl.iot.GetByAddress(string(req.Signer))
	if nil != err {
		return err
	}

	// log.Println("Key: ", getKey(iot.Id))
	_, err = impl.rClient.SetNX(
		context.TODO(), getKey(iot.Id), stateRaw, stateDuration,
	).Result()
	if nil != err {
		return err
	}
	// log.Println("Ok ? ", ok)

	return nil
}

func (impl *StateImpl) Get(req *domain.RStateGet,
) (*models.StateExtract, error) {
	var rs = &models.StateExtract{
		State:     models.StateInactived,
		CreatedAt: time.Now().Unix(),
	}

	data, err := impl.rClient.Get(context.TODO(), getKey(req.IotId)).Result()
	if nil != err {
		if err == redis.Nil {
			return rs, nil
		}

		return nil, err
	}
	if data == "" {
		return rs, nil
	}

	err = json.Unmarshal([]byte(data), rs)
	if nil != err {
		return nil, err
	}

	return rs, nil
}

func getKey(iotId int64) string {
	return fmt.Sprintf("iot_state_%d", iotId)
}
