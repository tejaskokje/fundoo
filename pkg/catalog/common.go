package catalog

import "errors"

var (
	ErrInvalidProduct       = errors.New("invalid product")
	ErrProductAlreadyExists = errors.New("product already exists")
)
