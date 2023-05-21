package utils

type ConstError string

var _ error = (*ConstError)(nil)

func (err ConstError) Error() string {
	return string(err)
}
