package core

type Error struct {
	Op   string
	Err  error
	Path string
}

func (e *Error) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }

func (e *Error) Unwrap() error { return e.Err }
