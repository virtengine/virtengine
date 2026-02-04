package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SupportResponseID uniquely identifies a support response.
type SupportResponseID struct {
	RequestID SupportRequestID `json:"request_id"`
	Sequence  uint64           `json:"sequence"`
}

// String returns the response ID string.
func (id SupportResponseID) String() string {
	return fmt.Sprintf("%s/response/%d", id.RequestID.String(), id.Sequence)
}

// Validate validates the response ID.
func (id SupportResponseID) Validate() error {
	if err := id.RequestID.Validate(); err != nil {
		return err
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// ParseSupportResponseID parses a response ID string.
func ParseSupportResponseID(value string) (SupportResponseID, error) {
	parts := strings.Split(value, "/")
	if len(parts) < 5 {
		return SupportResponseID{}, fmt.Errorf("invalid support response id: %s", value)
	}
	reqID, err := ParseSupportRequestID(strings.Join(parts[0:3], "/"))
	if err != nil {
		return SupportResponseID{}, err
	}
	if parts[3] != "response" {
		return SupportResponseID{}, fmt.Errorf("invalid response id segment: %s", value)
	}
	seq, err := strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		return SupportResponseID{}, fmt.Errorf("invalid response sequence: %w", err)
	}
	return SupportResponseID{
		RequestID: reqID,
		Sequence:  seq,
	}, nil
}

// SupportResponse represents an on-chain support response.
type SupportResponse struct {
	ID            SupportResponseID       `json:"id"`
	RequestID     SupportRequestID        `json:"request_id"`
	AuthorAddress string                  `json:"author_address"`
	IsAgent       bool                    `json:"is_agent"`
	Payload       EncryptedSupportPayload `json:"payload"`
	CreatedAt     time.Time               `json:"created_at"`
}

// NewSupportResponse creates a new response.
func NewSupportResponse(id SupportResponseID, author string, isAgent bool, payload EncryptedSupportPayload, now time.Time) *SupportResponse {
	return &SupportResponse{
		ID:            id,
		RequestID:     id.RequestID,
		AuthorAddress: author,
		IsAgent:       isAgent,
		Payload:       payload,
		CreatedAt:     now.UTC(),
	}
}

// Validate validates the response.
func (r *SupportResponse) Validate() error {
	if r == nil {
		return ErrInvalidSupportResponse.Wrap("response is nil")
	}
	if err := r.ID.Validate(); err != nil {
		return ErrInvalidSupportResponse.Wrapf("invalid id: %v", err)
	}
	if r.AuthorAddress == "" {
		return ErrInvalidAddress.Wrap("author address is required")
	}
	if r.RequestID.String() != r.ID.RequestID.String() {
		return ErrInvalidSupportResponse.Wrap("request id mismatch")
	}
	if err := r.Payload.Validate(); err != nil {
		return err
	}
	return nil
}
