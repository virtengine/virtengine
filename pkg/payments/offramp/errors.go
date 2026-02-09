package offramp

import "errors"

var (
	ErrAdapterUnavailable = errors.New("offramp adapter unavailable")
	ErrQuoteExpired       = errors.New("offramp quote expired")
	ErrInvalidRequest     = errors.New("offramp request invalid")
)
