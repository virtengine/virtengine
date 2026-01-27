package v1beta3

import (
	"errors"
)

var (
	// ErrNameDoesNotExist is the error when name does not exist
	ErrNameDoesNotExist = errors.New("Name does not exist")
	// ErrInvalidRequest is the error for invalid request
	ErrInvalidRequest = errors.New("Invalid request")
	// ErrDeploymentExists is the error when already deployment exists
	ErrDeploymentExists = errors.New("Deployment exists")
	// ErrDeploymentNotFound is the error when deployment not found
	ErrDeploymentNotFound = errors.New("Deployment not found")
	// ErrDeploymentClosed is the error when deployment is closed
	ErrDeploymentClosed = errors.New("Deployment closed")
	// ErrOwnerAcctMissing is the error for owner account missing
	ErrOwnerAcctMissing = errors.New("Owner account missing")
	// ErrInvalidGroups is the error when groups are empty
	ErrInvalidGroups = errors.New("Invalid groups")
	// ErrInvalidDeploymentID is the error for invalid deployment id
	ErrInvalidDeploymentID = errors.New("Invalid: deployment id")
	// ErrEmptyVersion is the error when version is empty
	ErrEmptyVersion = errors.New("Invalid: empty version")
	// ErrInvalidVersion is the error when version is invalid
	ErrInvalidVersion = errors.New("Invalid: deployment version")
	// ErrInternal is the error for internal error
	ErrInternal = errors.New("internal error")
	// ErrInvalidDeployment = is the error when deployment does not pass validation
	ErrInvalidDeployment = errors.New("Invalid deployment")
	// ErrInvalidGroupID is the error when already deployment exists
	ErrInvalidGroupID = errors.New("Deployment exists")
	// ErrGroupNotFound is the keeper's error for not finding a group
	ErrGroupNotFound = errors.New("Group not found")
	// ErrGroupClosed is the error when deployment is closed
	ErrGroupClosed = errors.New("Group already closed")
	// ErrGroupOpen is the error when deployment is closed
	ErrGroupOpen = errors.New("Group open")
	// ErrGroupPaused is the error when deployment is closed
	ErrGroupPaused = errors.New("Group paused")
	// ErrGroupNotOpen indicates the Group state has progressed beyond initial Open.
	ErrGroupNotOpen = errors.New("Group not open")
	// ErrGroupSpecInvalid indicates a GroupSpec has invalid configuration
	ErrGroupSpecInvalid = errors.New("GroupSpec invalid")

	// ErrInvalidDeposit indicates an invalid deposit
	ErrInvalidDeposit = errors.New("Deposit invalid")
	// ErrInvalidIDPath indicates an invalid ID path
	ErrInvalidIDPath = errors.New("ID path invalid")
	// ErrInvalidParam indicates an invalid chain parameter
	ErrInvalidParam = errors.New("parameter invalid")
	// ErrInvalidDeploymentDepositor indicates an invalid chain parameter
	ErrInvalidDeploymentDepositor = errors.New("invalid deployment depositor")
)
