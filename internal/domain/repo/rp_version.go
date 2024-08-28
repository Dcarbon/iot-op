package repo

import (
	"github.com/Dcarbon/go-shared/libs/container"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Version2Impl struct {
	version *container.SafeMap[int32, string] // Latest version of ota
	path    *container.SafeMap[int32, string] // Path to download version
}

type VersionImpl struct {
	db   *gorm.DB
	data map[int32]*models.Version
}

func NewVersion2Impl(initVersion map[int32]string, initPath map[int32]string) (*Version2Impl, error) {
	var vImpl = &Version2Impl{
		version: container.NewSafeMapFrom(initVersion),
		path:    container.NewSafeMapFrom(initPath),
	}

	return vImpl, nil
}

var Data map[int32]*models.Version

func NewVersionImpl(db *gorm.DB) (*VersionImpl, error) {
	err := db.AutoMigrate(
		&models.Version{},
	)
	if nil != err {
		return nil, err
	}
	var vImpl = &VersionImpl{
		db: db,
	}
	vImpl.InitVersion()
	Data = vImpl.data
	return vImpl, nil
}

func (vImpl *VersionImpl) InitVersion() {
	var versions = &[]models.Version{}
	result := make(map[int32]*models.Version)
	if err := vImpl.tblVersion().Find(&versions).Error; err != nil {
		vImpl.data = result
	}

	if len(*versions) > 0 {
		for _, val := range *versions {
			result[val.IotType] = &val
		}
	}
	vImpl.data = result
}

// func (vImpl *VersionImpl) SetVersion(req *domain.RVersionSet,
// ) error {
// 	vImpl.version.Set(req.IotType, req.Version)
// 	vImpl.path.Set(req.IotType, fmt.Sprintf("%s/static/iots/ota/%d/%s",
// 		utils.StringEnv(gutils.EXTERNAL_HOST, "http://localhost:4000"), req.IotType, req.Version))
// 	return nil
// }

// func (vImpl *VersionImpl) GetVersion(req *domain.RVersionGet,
// ) (string, string, error) {
// 	version, _ := vImpl.version.Get(req.IotType)
// 	path, _ := vImpl.path.Get(req.IotType)
// 	return version, path, nil
// }

func (vImpl *VersionImpl) SetVersion(req *domain.RVersionSet,
) error {
	version := models.Version{
		IotType: req.IotType,
		Version: req.Version,
		Path:    req.Path,
	}
	if err := vImpl.tblVersion().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "iot_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"version", "path"}),
	}).Create(&version).Error; err != nil {
		return err
	}
	Data[req.IotType] = &version
	return nil
}

func (vImpl *VersionImpl) GetVersion(req *domain.RVersionGet,
) (string, string, error) {
	value := vImpl.data[req.IotType]
	if value == nil {
		return "", "", nil
	}
	return value.Version, value.Path, nil
}

func (vImpl *VersionImpl) tblVersion() *gorm.DB {
	return vImpl.db.Table(models.TableNameVersion)
}
