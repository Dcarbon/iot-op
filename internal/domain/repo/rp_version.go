package repo

import (
	"github.com/Dcarbon/go-shared/libs/container"
	"github.com/Dcarbon/iot-op/internal/domain"
)

type VersionImpl struct {
	data *container.SafeMap[int32, string]
}

func NewVersionImpl(initData map[int32]string) (*VersionImpl, error) {
	var vImpl = &VersionImpl{
		data: container.NewSafeMap[int32, string](),
	}

	for k, v := range initData {
		vImpl.data.Set(k, v)
	}

	return vImpl, nil
}

func (vImpl *VersionImpl) SetVersion(req *domain.RVersionSet,
) error {
	vImpl.data.Set(req.IotType, req.Version)
	return nil
}

func (vImpl *VersionImpl) GetVersion(req *domain.RVersionGet,
) (string, error) {
	version, _ := vImpl.data.Get(req.IotType)
	return version, nil
}
