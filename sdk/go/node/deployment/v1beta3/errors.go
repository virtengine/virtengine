package v1beta3

import (
	"errors"
)

var (
	// ErrNameDoesNotExist is the error when name does not exist
	ErrNameDoesNotExist = errors.New("name does not exist")
	// ErrInvalidRequest is the error for invalid request
	ErrInvalidRequest = errors.New("invalid request")
	// ErrDeploymentExists is the error when already deployment exists
	ErrDeploymentExists = errors.New("deployment exists")
	// ErrDeploymentNotFound is the error when deployment not found
	ErrDeploymentNotFound = errors.New("deployment not found")
	// ErrDeploymentClosed is the error when deployment is closed
	ErrDeploymentClosed = errors.New("deployment closed")
	// ErrOwnerAcctMissing is the error for owner account missing
	ErrOwnerAcctMissing = errors.New("owner account missing")
	// ErrInvalidGroups is the error when groups are empty
	ErrInvalidGroups = errors.New("invalid groups")
	// ErrInvalidDeploymentID is the error for invalid deployment id
	ErrInvalidDeploymentID = errors.New("invalid: deployment id")
	// ErrEmptyVersion is the error when version is empty
	ErrEmptyVersion = errors.New("invalid: empty version")
	// ErrInvalidVersion is the error when version is invalid
	ErrInvalidVersion = errors.New("invalid: deployment version")
	// ErrInternal is the error for internal error
	ErrInternal = errors.New("internal error")
	// ErrInvalidDeployment = is the error when deployment does not pass validation
	ErrInvalidDeployment = errors.New("invalid deployment")
	// ErrInvalidGroupID is the error when already deployment exists
	ErrInvalidGroupID = errors.New("deployment exists")
	// ErrGroupNotFound is the keeper's error for not finding a group
	ErrGroupNotFound = errors.New("group not found")
	// ErrGroupClosed is the error when deployment is closed
	ErrGroupClosed = errors.New("group already closed")
	// ErrGroupOpen is the error when deployment is closed
	ErrGroupOpen = errors.New("group open")
	// ErrGroupPaused is the error when deployment is closed
	ErrGroupPaused = errors.New("group paused")
	// ErrGroupNotOpen indicates the Group state has progressed beyond initial Open.
	ErrGroupNotOpen = errors.New("group not open")
	// ErrGroupSpecInvalid indicates a GroupSpec has invalid configuration
	ErrGroupSpecInvalid = errors.New("groupSpec invalid")

	// ErrInvalidDeposit indicates an invalid deposit
	ErrInvalidDeposit = errors.New("deposit invalid")
	// ErrInvalidIDPath indicates an invalid ID path
	ErrInvalidIDPath = errors.New("ID path invalid")
	// ErrInvalidParam indicates an invalid chain parameter
	ErrInvalidParam = errors.New("parameter invalid")
	// ErrInvalidDeploymentDepositor indicates an invalid chain parameter
	ErrInvalidDeploymentDepositor = errors.New("invalid deployment depositor")
)
