package models

import "github.com/Dcarbon/go-shared/dmodels"

type State int

const (
	StateInactived = 1
	StateIdle      = 5
	StateActived   = 10
)

type Sensor struct {
	Type   dmodels.SensorType `json:"type"`
	State  State              `json:"state"`
	Metric map[string]float64 `json:"metric"`
}

type StateExtract struct {
	Signer    string            `json:"signer"`
	State     State             `json:"state"   `
	Info      map[string]string `json:"info"`
	Sensors   []*Sensor         `json:"sensors" `
	CreatedAt int64             `json:"createdAt"`
}
