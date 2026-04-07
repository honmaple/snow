package types

type (
	Store interface {
		Statics() Statics
		GetStatic(string) *Static
		GetStaticURL(string) string
	}
	Loader interface {
		Load() (Store, error)
	}
)
