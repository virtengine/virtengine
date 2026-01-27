// Package review provides aliases for the Review module types.
//
// VE-911: Provider public reviews
package review

import (
	"github.com/virtengine/virtengine/x/review/keeper"
	"github.com/virtengine/virtengine/x/review/types"
)

// Module name and store key constants
const (
	ModuleName  = types.ModuleName
	StoreKey    = types.StoreKey
	RouterKey   = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

// Type aliases for types package
type (
	// Keeper aliases
	Keeper = keeper.Keeper

	// Types aliases
	Review              = types.Review
	ReviewID            = types.ReviewID
	ReviewState         = types.ReviewState
	Reviews             = types.Reviews
	OrderReference      = types.OrderReference
	ProviderAggregation = types.ProviderAggregation
	RatingDistribution  = types.RatingDistribution
	Params              = types.Params
	GenesisState        = types.GenesisState

	// Message types
	MsgSubmitReview = types.MsgSubmitReview
	MsgDeleteReview = types.MsgDeleteReview
	MsgUpdateParams = types.MsgUpdateParams
)

// Review state constants
const (
	ReviewStateUnspecified = types.ReviewStateUnspecified
	ReviewStateActive      = types.ReviewStateActive
	ReviewStateHidden      = types.ReviewStateHidden
	ReviewStateDeleted     = types.ReviewStateDeleted
)

// Rating constants
const (
	MinRating           = types.MinRating
	MaxRating           = types.MaxRating
	MinReviewTextLength = types.MinReviewTextLength
	MaxReviewTextLength = types.MaxReviewTextLength
)

// Function aliases
var (
	// Keeper constructors
	NewKeeper = keeper.NewKeeper

	// Types constructors
	NewReview              = types.NewReview
	NewProviderAggregation = types.NewProviderAggregation
	DefaultParams          = types.DefaultParams
	DefaultGenesisState    = types.DefaultGenesisState

	// Message constructors
	NewMsgSubmitReview = types.NewMsgSubmitReview
	NewMsgDeleteReview = types.NewMsgDeleteReview
	NewMsgUpdateParams = types.NewMsgUpdateParams

	// Codec registration
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec
	RegisterInterfaces       = types.RegisterInterfaces
)

// Error aliases
var (
	ErrInvalidRating       = types.ErrInvalidRating
	ErrInvalidReviewText   = types.ErrInvalidReviewText
	ErrOrderNotFound       = types.ErrOrderNotFound
	ErrOrderNotCompleted   = types.ErrOrderNotCompleted
	ErrUnauthorizedReviewer = types.ErrUnauthorizedReviewer
	ErrDuplicateReview     = types.ErrDuplicateReview
	ErrReviewNotFound      = types.ErrReviewNotFound
	ErrProviderNotFound    = types.ErrProviderNotFound
	ErrInvalidReviewID     = types.ErrInvalidReviewID
	ErrInvalidAddress      = types.ErrInvalidAddress
	ErrContentHashMismatch = types.ErrContentHashMismatch
	ErrReviewTextTooLong   = types.ErrReviewTextTooLong
	ErrReviewTextTooShort  = types.ErrReviewTextTooShort
)

// Event type aliases
const (
	EventTypeReviewSubmitted    = types.EventTypeReviewSubmitted
	EventTypeReviewUpdated      = types.EventTypeReviewUpdated
	EventTypeReviewDeleted      = types.EventTypeReviewDeleted
	EventTypeAggregationUpdated = types.EventTypeAggregationUpdated
)

// Attribute key aliases
const (
	AttributeKeyReviewID        = types.AttributeKeyReviewID
	AttributeKeyOrderID         = types.AttributeKeyOrderID
	AttributeKeyProviderAddress = types.AttributeKeyProviderAddress
	AttributeKeyReviewerAddress = types.AttributeKeyReviewerAddress
	AttributeKeyRating          = types.AttributeKeyRating
	AttributeKeyContentHash     = types.AttributeKeyContentHash
	AttributeKeyAverageRating   = types.AttributeKeyAverageRating
	AttributeKeyTotalReviews    = types.AttributeKeyTotalReviews
)
