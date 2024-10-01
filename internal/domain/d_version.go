package domain

type RVersionSet struct {
	IotType int32
	Version string
	Path    string
}

type RVersionGet struct {
	IotType int32
	Version *string
}

type IVersion interface {
	SetVersion(*RVersionSet) error
	GetVersion(*RVersionGet) (string, string, error)
}
