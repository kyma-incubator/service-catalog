package mutation

type MutateError struct {
	errMsg  string
	errCode int32
}

func NewMutateError(msg string, code int32) *MutateError {
	return &MutateError{
		errMsg:  msg,
		errCode: code,
	}
}

func (m *MutateError) Error() string {
	return m.errMsg
}

func (m *MutateError) Code() int32 {
	return m.errCode
}
