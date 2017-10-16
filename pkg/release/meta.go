package release

const (
	// MetaFileName is the name of metadata file.
	MetaFileName = "meta.yml"
)

// Meta represents metadata of a release.
type Meta struct {
	Binaries []*Binary `yaml:"binaries"`
}

// NewMeta returns a Meta object.
func NewMeta(bins []*Binary) *Meta {
	return &Meta{Binaries: bins}
}
