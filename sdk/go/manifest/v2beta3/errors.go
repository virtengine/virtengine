package v2beta3

import (
	"errors"
)

var (
	ErrInvalidManifest         = errors.New("invalid manifest")
	ErrManifestCrossValidation = errors.New("manifest cross-validation error")
)
