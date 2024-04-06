package repo

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/go-shared/ecodes"
	"github.com/Dcarbon/go-shared/gutils"
	"github.com/Dcarbon/go-shared/libs/esign"
	"github.com/Dcarbon/go-shared/svc"
	"github.com/Dcarbon/iot-op/internal/domain"
	"github.com/Dcarbon/iot-op/internal/models"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type SolMintImpl struct {
	db   *gorm.DB
	iiot svc.IIotInfo
}

func NewSolanaMinterImpl(db *gorm.DB, iiot svc.IIotInfo,
) (*SolMintImpl, error) {

	err := db.AutoMigrate(
		&models.MintSign{},
		&models.Minted{},
	)
	if nil != err {
		return nil, err
	}

	var ip = &SolMintImpl{
		db:   db,
		iiot: iiot,
	}
	return ip, nil
}

func (ip *SolMintImpl) Mint(req *domain.RMinterMint,
) error {
	if req.Nonce <= 0 {
		return dmodels.ErrInvalidNonce()
	}

	newAmount, e1 := dmodels.NewBigNumberFromHex(req.Amount)
	if nil != e1 {
		return e1
	}

	req.Iot = strings.ToLower(req.Iot)
	iot, e1 := ip.iiot.GetByAddress(req.Iot)
	if nil != e1 {
		return e1
	}

	if iot.Status < dmodels.DeviceStatusRegister {
		return dmodels.NewError(ecodes.IOTNotAllowed, "IOT is not allow")
	}

	var mint = &models.MintSign{
		Id:        0,
		Nonce:     req.Nonce,
		Amount:    req.Amount,
		IotId:     iot.Id,
		Iot:       req.Iot,
		Signed:    req.Signed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	e1 = mint.VerifySolana()
	if nil != e1 {
		return e1
	}

	var latest = make([]*models.MintSign, 0, 1)
	e1 = ip.tblSign().
		Where("iot = ?", mint.Iot).
		Order("created_at desc").
		Limit(1).
		Find(&latest).Error
	if nil != e1 {
		return dmodels.ParsePostgresError("", e1)
	}

	if len(latest) == 0 {
		latest = append(latest, &models.MintSign{})
	}

	if latest[0].Nonce == mint.Nonce || latest[0].Nonce+1 == mint.Nonce {
		oldAmount, e1 := dmodels.NewBigNumberFromHex(latest[0].Amount)
		if nil != e1 {
			oldAmount = dmodels.NewBigNumber(0)
		}

		var incAmount = big.NewInt(0).Sub(newAmount.Int, oldAmount.Int)
		var minted = &models.Minted{
			Id:     uuid.NewV4().String(),
			IotId:  iot.Id,
			Carbon: incAmount.Int64(),
		}

		return ip.db.Transaction(func(dbTx *gorm.DB) error {
			if latest[0].Nonce+1 == mint.Nonce {
				err := dbTx.Table(models.TableNameMintSign).Create(mint).Error
				if nil != err {
					return dmodels.ParsePostgresError("", err)
				}
			} else {
				err := dbTx.Table(models.TableNameMintSign).
					Where("id = ?", latest[0].Id).
					Updates(map[string]interface{}{
						"iot_id":     mint.IotId,
						"nonce":      mint.Nonce,
						"amount":     mint.Amount,
						"signed":     mint.Signed,
						"updated_at": time.Now(),
					}).Error
				if nil != err {
					dmodels.ParsePostgresError("", err)
				}
			}

			err := dbTx.Table(models.TableNameMinted).Create(minted).Error
			if nil != err {
				return dmodels.ParsePostgresError("", err)
			}
			return nil
		})

	}
	return dmodels.ErrInvalidNonce()
}

func (ip *SolMintImpl) GetSigns(req *domain.RMinterGetSigns,
) ([]*models.MintSign, error) {
	var iot, err = ip.iiot.GetById(req.IotId)
	if nil != err {
		return nil, err
	}

	var signeds = make([]*models.MintSign, 0)
	var query = ip.tblSign().
		Where(
			"updated_at > ? AND updated_at < ? AND iot = ?",
			time.Unix(req.From, 0), time.Unix(req.To, 0), iot.Address,
		)

	if req.Sort > 0 {
		query = query.Order("updated_at desc")
	} else {
		query = query.Order("updated_at asc")
	}

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	err = query.Find(&signeds).Error
	if nil != err {
		return nil, dmodels.ParsePostgresError("Get mint sign", err)
	}
	return signeds, nil
}

func (ip *SolMintImpl) GetSignLatest(req *domain.RMinterGetSignLatest,
) (*models.MintSign, error) {
	var iot, err = ip.iiot.GetById(req.IotId)
	if nil != err {
		return nil, err
	}

	var signed = &models.MintSign{}
	var query = ip.tblSign().Where("iot = ?", iot.Address).Order("updated_at desc")
	err = query.Find(&signed).Error
	if nil != err {
		return nil, dmodels.ParsePostgresError("Get mint sign", err)
	}
	return signed, nil
}

func (ip *SolMintImpl) GetMinted(req *domain.RMinterGetMinted,
) ([]*models.Minted, error) {
	req.Normalize()

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

		query = ip.tblMinted().Raw(
			fmt.Sprintf(
				`WITH cte_minted as (%s) SELECT * FROM cte_minted`,
				ip.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Table(models.TableNameMinted).
						Select(
							fmt.Sprintf("date_trunc('%s', \"created_at\", ?) as created_at, sum(carbon) as carbon", group),
							tz,
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
		query = ip.tblMinted().
			Select("created_at, carbon").
			Where(
				"created_at > ? AND created_at < ? AND iot_id = ? ",
				time.Unix(req.From, 0), time.Unix(req.To, 0), req.IotId,
			).
			Offset(req.Skip).
			Limit(int(req.Limit)).
			Order("created_at " + req.Sort.String())
	}

	var data = make([]*models.Minted, 0)
	var err = query.Debug().Find(&data).Error
	if nil != err {
		return nil, dmodels.ParsePostgresError("", err)
	}
	return data, nil
}

func (ip *SolMintImpl) GetSeparator() (*esign.TypedDataDomain, error) {
	// return ip.dMinter.GetDomain(), nil
	return nil, nil
}

func (ip *SolMintImpl) tblSign() *gorm.DB {
	return ip.db.Table(models.TableNameMintSign)
}

func (ip *SolMintImpl) tblMinted() *gorm.DB {
	return ip.db.Table(models.TableNameMinted)
}