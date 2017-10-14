package meta

type Binary struct {
	Name      string `yaml:"name"`
	Checksum  string `yaml:"checksum"`
	Timestamp string `yaml:"timestamp"`
	Version   string `yaml:"version,omitempty"`
}

type Meta struct {
	Binaries []*Binary `yaml:"binaries"`
}

func New(b *Binary) *Meta {
	return &Meta{Binaries: []*Binary{b}}
}

func (m *Meta) AppendBinary(b *Binary) {
	m.Binaries = append(m.Binaries, b)
}
