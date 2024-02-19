package domain

import (
	"github.com/Dcarbon/go-shared/dmodels"
	"github.com/Dcarbon/iot-op/internal/models"
)

// Iot state update
type RStateUpdate struct {
	dmodels.Signature
}

type RStateGet struct {
	IotId int64
}

// Update state
type IState interface {
	Update(*RStateUpdate) error
	Get(*RStateGet) (*models.StateExtract, error)
}
