package v1

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errNameDoesNotExist uint32 = iota + 1
	errInvalidRequest
	errDeploymentExists
	errDeploymentNotFound
	errDeploymentClosed
	errOwnerAcctMissing
	errInvalidGroups
	errInvalidDeploymentID
	errEmptyHash
	errInvalidHash
	errInternal
	errInvalidDeployment
	errInvalidGroupID
	errGroupNotFound
	errGroupClosed
	errGroupOpen
	errGroupPaused
	errGroupNotOpen
	errGroupSpec
	errInvalidDeposit
	errInvalidIDPath
	errInvalidParam
	errInvalidEscrowID
)

var (
	// ErrNameDoesNotExist is the error when name does not exist
	ErrNameDoesNotExist = sdkerrors.RegisterWithGRPCCode(ModuleName, errNameDoesNotExist, codes.NotFound, "Name does not exist")
	// ErrInvalidRequest is the error for invalid request
	ErrInvalidRequest = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidRequest, codes.InvalidArgument, "Invalid request")
	// ErrDeploymentExists is the error when already deployment exists
	ErrDeploymentExists = sdkerrors.RegisterWithGRPCCode(ModuleName, errDeploymentExists, codes.AlreadyExists, "Deployment exists")
	// ErrDeploymentNotFound is the error when deployment not found
	ErrDeploymentNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errDeploymentNotFound, codes.NotFound, "Deployment not found")
	// ErrDeploymentClosed is the error when deployment is closed
	ErrDeploymentClosed = sdkerrors.RegisterWithGRPCCode(ModuleName, errDeploymentClosed, codes.FailedPrecondition, "Deployment closed")
	// ErrOwnerAcctMissing is the error for owner account missing
	ErrOwnerAcctMissing = sdkerrors.RegisterWithGRPCCode(ModuleName, errOwnerAcctMissing, codes.InvalidArgument, "Owner account missing")
	// ErrInvalidGroups is the error when groups are empty
	ErrInvalidGroups = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidGroups, codes.InvalidArgument, "Invalid groups")
	// ErrInvalidDeploymentID is the error for invalid deployment id
	ErrInvalidDeploymentID = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidDeploymentID, codes.InvalidArgument, "Invalid: deployment id")
	// ErrEmptyHash is the error when version is empty
	ErrEmptyHash = sdkerrors.RegisterWithGRPCCode(ModuleName, errEmptyHash, codes.InvalidArgument, "Invalid: empty hash")
	// ErrInvalidHash is the error when version is invalid
	ErrInvalidHash = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidHash, codes.InvalidArgument, "Invalid: deployment hash")
	// ErrInternal is the error for internal error
	ErrInternal = sdkerrors.RegisterWithGRPCCode(ModuleName, errInternal, codes.Internal, "internal error")
	// ErrInvalidDeployment = is the error when deployment does not pass validation
	ErrInvalidDeployment = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidDeployment, codes.InvalidArgument, "Invalid deployment")
	// ErrInvalidGroupID is the error when the deployment's group ID is invalid
	ErrInvalidGroupID = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidGroupID, codes.InvalidArgument, "Invalid deployment's group ID")
	// ErrGroupNotFound is the keeper's error for not finding a group
	ErrGroupNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errGroupNotFound, codes.NotFound, "Group not found")
	// ErrGroupClosed is the error when group is closed
	ErrGroupClosed = sdkerrors.RegisterWithGRPCCode(ModuleName, errGroupClosed, codes.FailedPrecondition, "Group already closed")
	// ErrGroupOpen is the error when group is open
	ErrGroupOpen = sdkerrors.RegisterWithGRPCCode(ModuleName, errGroupOpen, codes.FailedPrecondition, "Group open")
	// ErrGroupPaused is the error when group is paused
	ErrGroupPaused = sdkerrors.RegisterWithGRPCCode(ModuleName, errGroupPaused, codes.FailedPrecondition, "Group paused")
	// ErrGroupNotOpen indicates the Group state has progressed beyond initial Open.
	ErrGroupNotOpen = sdkerrors.RegisterWithGRPCCode(ModuleName, errGroupNotOpen, codes.FailedPrecondition, "Group not open")
	// ErrGroupSpecInvalid indicates a GroupSpec has invalid configuration
	ErrGroupSpecInvalid = sdkerrors.RegisterWithGRPCCode(ModuleName, errGroupSpec, codes.InvalidArgument, "GroupSpec invalid")
	// ErrInvalidDeposit indicates an invalid deposit
	ErrInvalidDeposit = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidDeposit, codes.InvalidArgument, "Deposit invalid")
	// ErrInvalidIDPath indicates an invalid ID path
	ErrInvalidIDPath = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidIDPath, codes.InvalidArgument, "ID path invalid")
	// ErrInvalidParam indicates an invalid chain parameter
	ErrInvalidParam = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidParam, codes.InvalidArgument, "parameter invalid")
	// ErrInvalidEscrowID indicates an invalid escrow ID
	ErrInvalidEscrowID = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidEscrowID, codes.InvalidArgument, "invalid escrow id")
)
