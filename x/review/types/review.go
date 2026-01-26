// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Review type definitions
// This file defines the Review type with star ratings (1-5), text content,
// verified order links, and on-chain hash for content integrity.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Rating constraints
const (
	// MinRating is the minimum rating value
	MinRating = 1

	// MaxRating is the maximum rating value
	MaxRating = 5

	// MinReviewTextLength is the minimum review text length
	MinReviewTextLength = 10

	// MaxReviewTextLength is the maximum review text length
	MaxReviewTextLength = 2000
)

// ReviewState represents the state of a review
type ReviewState uint8

const (
	// ReviewStateUnspecified represents an unspecified review state
	ReviewStateUnspecified ReviewState = 0

	// ReviewStateActive indicates the review is active and visible
	ReviewStateActive ReviewState = 1

	// ReviewStateHidden indicates the review is hidden (by moderator)
	ReviewStateHidden ReviewState = 2

	// ReviewStateDeleted indicates the review has been deleted
	ReviewStateDeleted ReviewState = 3
)

// ReviewStateNames maps review states to human-readable names
var ReviewStateNames = map[ReviewState]string{
	ReviewStateUnspecified: "unspecified",
	ReviewStateActive:      "active",
	ReviewStateHidden:      "hidden",
	ReviewStateDeleted:     "deleted",
}

// String returns the string representation of a ReviewState
func (s ReviewState) String() string {
	if name, ok := ReviewStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the review state is valid
func (s ReviewState) IsValid() bool {
	return s >= ReviewStateActive && s <= ReviewStateDeleted
}

// IsVisible returns true if the review should be publicly visible
func (s ReviewState) IsVisible() bool {
	return s == ReviewStateActive
}

// ReviewID is the unique identifier for a review
type ReviewID struct {
	// ProviderAddress is the provider being reviewed
	ProviderAddress string `json:"provider_address"`

	// Sequence is the provider-scoped sequential review number
	Sequence uint64 `json:"sequence"`
}

// String returns the string representation of the review ID
func (id ReviewID) String() string {
	return fmt.Sprintf("%s/review/%d", id.ProviderAddress, id.Sequence)
}

// Validate validates the review ID
func (id ReviewID) Validate() error {
	if id.ProviderAddress == "" {
		return fmt.Errorf("provider address is required")
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// OrderReference contains the verified order link information
type OrderReference struct {
	// OrderID is the unique order identifier string
	OrderID string `json:"order_id"`

	// CustomerAddress is the customer who placed the order
	CustomerAddress string `json:"customer_address"`

	// ProviderAddress is the provider who fulfilled the order
	ProviderAddress string `json:"provider_address"`

	// CompletedAt is when the order was completed
	CompletedAt time.Time `json:"completed_at"`

	// OrderHash is the hash of the order for verification
	OrderHash string `json:"order_hash"`
}

// Validate validates the order reference
func (ref *OrderReference) Validate() error {
	if ref.OrderID == "" {
		return ErrInvalidOrderID.Wrap("order ID is required")
	}
	if ref.CustomerAddress == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}
	if ref.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}
	if ref.CompletedAt.IsZero() {
		return fmt.Errorf("completed_at is required")
	}
	return nil
}

// Review represents a provider review with star rating and text content
type Review struct {
	// ID is the unique review identifier
	ID ReviewID `json:"id"`

	// ReviewerAddress is the blockchain address of the reviewer
	ReviewerAddress string `json:"reviewer_address"`

	// ProviderAddress is the provider being reviewed
	ProviderAddress string `json:"provider_address"`

	// OrderRef contains the verified order reference
	OrderRef OrderReference `json:"order_ref"`

	// Rating is the star rating (1-5)
	Rating uint8 `json:"rating"`

	// Text is the review text content
	Text string `json:"text"`

	// ContentHash is the SHA256 hash of the review content (rating + text)
	// stored on-chain for integrity verification
	ContentHash string `json:"content_hash"`

	// State is the current review state
	State ReviewState `json:"state"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updated_at"`

	// BlockHeight is the block height when the review was created
	BlockHeight int64 `json:"block_height"`

	// ModerationReason contains reason if review was hidden/deleted by moderator
	ModerationReason string `json:"moderation_reason,omitempty"`

	// ModeratorAddress is the moderator who took action (if any)
	ModeratorAddress string `json:"moderator_address,omitempty"`
}

// NewReview creates a new review with required fields
func NewReview(
	id ReviewID,
	reviewerAddress string,
	providerAddress string,
	orderRef OrderReference,
	rating uint8,
	text string,
) (*Review, error) {
	now := time.Now().UTC()

	review := &Review{
		ID:              id,
		ReviewerAddress: reviewerAddress,
		ProviderAddress: providerAddress,
		OrderRef:        orderRef,
		Rating:          rating,
		Text:            text,
		State:           ReviewStateActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Compute and set content hash
	review.ContentHash = review.ComputeContentHash()

	return review, review.Validate()
}

// Validate validates the review
func (r *Review) Validate() error {
	if err := r.ID.Validate(); err != nil {
		return ErrInvalidReviewID.Wrapf("invalid review ID: %v", err)
	}

	if r.ReviewerAddress == "" {
		return ErrInvalidAddress.Wrap("reviewer address is required")
	}

	if r.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if err := r.OrderRef.Validate(); err != nil {
		return err
	}

	// Validate rating is between 1-5
	if r.Rating < MinRating || r.Rating > MaxRating {
		return ErrInvalidRating.Wrapf("rating %d is outside range [%d, %d]", r.Rating, MinRating, MaxRating)
	}

	// Validate review text length
	if len(r.Text) < MinReviewTextLength {
		return ErrReviewTextTooShort.Wrapf("minimum length is %d characters", MinReviewTextLength)
	}
	if len(r.Text) > MaxReviewTextLength {
		return ErrReviewTextTooLong.Wrapf("maximum length is %d characters", MaxReviewTextLength)
	}

	if !r.State.IsValid() {
		return fmt.Errorf("invalid review state: %s", r.State)
	}

	return nil
}

// ComputeContentHash computes the SHA256 hash of the review content
func (r *Review) ComputeContentHash() string {
	h := sha256.New()
	// Include rating, text, order ID, and provider for integrity
	h.Write([]byte(fmt.Sprintf("%d:%s:%s:%s",
		r.Rating,
		r.Text,
		r.OrderRef.OrderID,
		r.ProviderAddress,
	)))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyContentHash verifies the content hash matches the current content
func (r *Review) VerifyContentHash() error {
	expectedHash := r.ComputeContentHash()
	if r.ContentHash != expectedHash {
		return ErrContentHashMismatch.Wrapf("expected %s, got %s", expectedHash, r.ContentHash)
	}
	return nil
}

// IsVisibleToPublic returns true if the review should be shown publicly
func (r *Review) IsVisibleToPublic() bool {
	return r.State.IsVisible()
}

// Hide hides the review with a moderation reason
func (r *Review) Hide(moderatorAddr, reason string) error {
	if r.State == ReviewStateDeleted {
		return fmt.Errorf("cannot hide deleted review")
	}
	r.State = ReviewStateHidden
	r.ModeratorAddress = moderatorAddr
	r.ModerationReason = reason
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// Delete marks the review as deleted
func (r *Review) Delete(moderatorAddr, reason string) error {
	r.State = ReviewStateDeleted
	r.ModeratorAddress = moderatorAddr
	r.ModerationReason = reason
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// Restore restores a hidden review to active
func (r *Review) Restore() error {
	if r.State != ReviewStateHidden {
		return fmt.Errorf("can only restore hidden reviews")
	}
	r.State = ReviewStateActive
	r.ModeratorAddress = ""
	r.ModerationReason = ""
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// Reviews is a slice of Review
type Reviews []Review

// Active returns only active reviews
func (reviews Reviews) Active() Reviews {
	result := make(Reviews, 0)
	for _, r := range reviews {
		if r.State == ReviewStateActive {
			result = append(result, r)
		}
	}
	return result
}

// ByProvider returns reviews for a specific provider
func (reviews Reviews) ByProvider(providerAddress string) Reviews {
	result := make(Reviews, 0)
	for _, r := range reviews {
		if r.ProviderAddress == providerAddress {
			result = append(result, r)
		}
	}
	return result
}

// ByReviewer returns reviews by a specific reviewer
func (reviews Reviews) ByReviewer(reviewerAddress string) Reviews {
	result := make(Reviews, 0)
	for _, r := range reviews {
		if r.ReviewerAddress == reviewerAddress {
			result = append(result, r)
		}
	}
	return result
}

// AverageRating calculates the average rating of active reviews
func (reviews Reviews) AverageRating() float64 {
	active := reviews.Active()
	if len(active) == 0 {
		return 0.0
	}

	var total int64
	for _, r := range active {
		total += int64(r.Rating)
	}
	return float64(total) / float64(len(active))
}
