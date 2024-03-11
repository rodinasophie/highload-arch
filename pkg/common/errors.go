package common

import (
	"github.com/pkg/errors"
)

var ErrUserNotFound = errors.Errorf("User not found")
var ErrTokenNotFound = errors.Errorf("Token not found")
var ErrTokenExpired = errors.Errorf("Token has expired")
var ErrTokenInvalid = errors.Errorf("Token is invalid")
var ErrPasswordInvalid = errors.Errorf("Password is invalid")
var ErrPostNotFound = errors.Errorf("Post not found")
var ErrRequestNotAuthorized = errors.Errorf("Request not authorized")
