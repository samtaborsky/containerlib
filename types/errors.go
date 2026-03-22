// SPDX-License-Identifier: MIT

package types

import (
	"errors"
)

var (
	ErrContainerNotFound = errors.New("container not found")
	ErrImageNotFound     = errors.New("image not found")
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("container state conflict")
	ErrPermissionsDenied = errors.New("permission denied")
	ErrConnectionFailed  = errors.New("connection failed")
	ErrInternal          = errors.New("internal error")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
)
