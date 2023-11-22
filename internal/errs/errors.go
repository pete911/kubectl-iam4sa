package errs

type ErrNotFound struct {
	msg string
}

func NewErrNotFound(msg string) *ErrNotFound {
	return &ErrNotFound{msg: msg}
}

func (e *ErrNotFound) Error() string {
	return e.msg
}
