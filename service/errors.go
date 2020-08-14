package service

type ErrNotValidRequest struct {
	Reason string
}

func (e ErrNotValidRequest) Error() string {
	return e.Reason
}
