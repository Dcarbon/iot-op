package repo

import (
	"github.com/Dcarbon/go-shared/libs/container"
	"github.com/Dcarbon/iot-op/internal/domain"
)

type VersionImpl struct {
	version *container.SafeMap[int32, string] // Latest version of ota
	path    *container.SafeMap[int32, string] // Path to download version
}

func NewVersionImpl(initVersion map[int32]string, initPath map[int32]string) (*VersionImpl, error) {
	var vImpl = &VersionImpl{
		version: container.NewSafeMapFrom(initVersion),
		path:    container.NewSafeMapFrom(initPath),
	}

	return vImpl, nil
}

func (vImpl *VersionImpl) SetVersion(req *domain.RVersionSet,
) error {
	vImpl.version.Set(req.IotType, req.Version)
	return nil
}

func (vImpl *VersionImpl) GetVersion(req *domain.RVersionGet,
) (string, string, error) {
	version, _ := vImpl.version.Get(req.IotType)
	path, _ := vImpl.path.Get(req.IotType)
	return version, path, nil
}
