package repo

import (
	"fmt"
	"strings"
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/gutils"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type AVMImpl struct {
	db  *gorm.DB
	iot svc.IIotInfo
}

func NewAVMImpl(db *gorm.DB, iot svc.IIotInfo) (*AVMImpl, error) {
	err := db.AutoMigrate(&models.AVM{})
	if nil != err {
		return nil, err
	}

	var aimpl = &AVMImpl{
		db:  db,
		iot: iot,
	}

	return aimpl, nil
}

func (aimpl *AVMImpl) Create(req *domain.RAVMCreate,
) (*models.AVM, error) {
	err := req.Verify()
	if nil != err {
		return nil, err
	}

	x, err := req.Extract()
	if nil != err {
		return nil, err
	}

	iot, err := aimpl.iot.GetByAddress(string(req.Signer))
	if nil != err {
		return nil, err
	}

	avm := &models.AVM{
		Id:        uuid.NewV4().String(),
		IotId:     iot.Id,
		Signed:    req.Signed,
		Data:      req.Data,
		Volume:    x.Volume,
		CreatedAt: time.Unix(x.From, 0),
	}
	err = aimpl.tblAVM().Create(avm).Error
	if nil != err {
		return nil, err
	}
	return avm, nil
}

func (aimpl *AVMImpl) GetList(req *domain.RAVMGetList,
) ([]*models.AVM, error) {
	if req.To == 0 {
		req.To = time.Now().Unix()
	}

	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 100
	}

	var query *gorm.DB
	if req.Interval > 0 {
		var tz = "Asia/Ho_Chi_Minh"
		var group = req.Interval.String()
		if group == "" {
			return nil, gutils.ErrBadRequest("Invalid interval param")
		}
		if req.Interval == 2 {
			group = "month"
		}

		query = aimpl.tblAVM().Raw(
			fmt.Sprintf(
				`WITH cte_avm as (%s) SELECT * FROM cte_avm`,
				aimpl.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Table(models.TableNameAVM).
						Select(
							fmt.Sprintf(
								"date_trunc('%s', \"created_at\", '%s') as created_at, sum(volume) as volume",
								group, tz,
							),
						).
						Where(
							"created_at > ? AND created_at < ? AND iot_id = ? ",
							time.Unix(req.From, 0), time.Unix(req.To, 0), req.IotId,
						).
						Group(fmt.Sprintf("date_trunc('%s', \"created_at\", '%s')", group, tz)).
						Order("created_at " + req.Sort.String()).
						Offset(req.Skip).
						Limit(int(req.Limit)).
						Find(&struct{}{})
				}),
			),
		)
	} else {
		var selectFields = []string{"volume", "created_at"}
		if req.Full {
			selectFields = append(selectFields, "id", "signed", "data", "iot_id")
		}

		query = aimpl.tblAVM().
			Select(strings.Join(selectFields, ", ")).
			Where(
				"created_at > ? AND created_at < ? AND iot_id = ? ",
				time.Unix(req.From, 0), time.Unix(req.To, 0), req.IotId,
			).
			Offset(req.Skip).
			Limit(int(req.Limit)).
			Order("created_at " + req.Sort.String())
	}

	var data = make([]*models.AVM, 0)
	var err = query.Find(&data).Error
	if nil != err {
		return nil, dmodels.ParsePostgresError("", err)
	}

	return data, nil
}

func (aimpl *AVMImpl) tblAVM() *gorm.DB {
	return aimpl.db.Table(models.TableNameAVM)
}
