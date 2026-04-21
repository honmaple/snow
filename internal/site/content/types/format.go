package types

type (
	Format struct {
		Name      string
		Path      string
		Template  string
		Permalink string
	}
	Formats []*Format
)

func (fs Formats) Find(name string) *Format {
	for _, f := range fs {
		if f.Name == name {
			return f
		}
	}
	return nil
}
