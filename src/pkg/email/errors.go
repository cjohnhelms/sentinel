package email

import "errors"

var (
	ErrEmail        = errors.New("email failed")
	ErrBodyGenerate = errors.New("failed to generate email body")
)
