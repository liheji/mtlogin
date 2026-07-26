package conf

import "errors"

var (
	ErrRead    = errors.New("config read failed")
	ErrParse   = errors.New("config parse failed")
	ErrInvalid = errors.New("config invalid")
)
