package model

import "fmt"

var (
	ErrPendingImage   = fmt.Errorf("image is pending")
	ErrImageNotExists = fmt.Errorf("image not exists")
)
