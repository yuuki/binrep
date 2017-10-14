package release

type Meta struct {
	Binaries []*Binary `yaml:"binaries"`
}

func NewMeta(bins []*Binary) *Meta {
	return &Meta{Binaries: bins}
}
