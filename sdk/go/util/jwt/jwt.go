package jwt

import (
	"bytes"
	"encoding/csv"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang-jwt/jwt/v5"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1"
)

var (
	ErrJWTValidation = fmt.Errorf("jwt: validation error")
)

type AccessType string
type PermissionScope string

const (
	AccessTypeFull     AccessType = "full"
	AccessTypeGranular AccessType = "granular"
	AccessTypeScoped   AccessType = "scoped"
	AccessTypeNone     AccessType = "none"
)

const (
	PermissionScopeSendManifest    PermissionScope = "send-manifest"
	PermissionScopeGetManifest     PermissionScope = "get-manifest"
	PermissionScopeLogs            PermissionScope = "logs"
	PermissionScopeShell           PermissionScope = "shell"
	PermissionScopeEvents          PermissionScope = "events"
	PermissionScopeStatus          PermissionScope = "status"
	PermissionScopeRestart         PermissionScope = "restart"
	PermissionScopeHostnameMigrate PermissionScope = "hostname-migrate"
	PermissionScopeIPMigrate       PermissionScope = "ip-migrate"
)

var (
	validScopes = map[PermissionScope]bool{
		PermissionScopeSendManifest:    true,
		PermissionScopeGetManifest:     true,
		PermissionScopeLogs:            true,
		PermissionScopeShell:           true,
		PermissionScopeEvents:          true,
		PermissionScopeRestart:         true,
		PermissionScopeStatus:          true,
		PermissionScopeHostnameMigrate: true,
		PermissionScopeIPMigrate:       true,
	}
)

func (s PermissionScope) String() string {
	return string(s)
}

// GetSupportedScopes returns the list of permission scopes recognised by the
// JWT package.
func GetSupportedScopes() PermissionScopes {
	out := make(PermissionScopes, 0, len(validScopes))

	for k := range validScopes {
		out = append(out, k)
	}

	return out
}

type PermissionScopes []PermissionScope

type PermissionDeployment struct {
	Scope    PermissionScopes `json:"scope"`
	DSeq     uint64           `json:"dseq,omitempty"`
	GSeq     uint32           `json:"gseq,omitempty"`
	OSeq     uint32           `json:"oseq,omitempty"`
	Services []string         `json:"services,omitempty"`
}

type Permission struct {
	Provider    sdk.AccAddress         `json:"provider"`
	Access      AccessType             `json:"access"`
	Scope       PermissionScopes       `json:"scope,omitempty"`
	Deployments []PermissionDeployment `json:"deployments,omitempty"`
}

type Permissions []Permission

type Claims struct {
	jwt.RegisteredClaims
	Version string `json:"version"`
	Leases  Leases `json:"leases"`

	iss sdk.AccAddress
}

var _ jwt.Claims = (*Claims)(nil)

type Leases struct {
	Access      AccessType       `json:"access"`
	Scope       PermissionScopes `json:"scope,omitempty"`
	Permissions Permissions      `json:"permissions,omitempty"`
}

func (c PermissionScopes) Has(scope PermissionScope) bool {
	for i := range c {
		if c[i] == scope {
			return true
		}
	}

	return false
}

func (c *PermissionScopes) UnmarshalCSV(val string) error {
	rd := csv.NewReader(bytes.NewBufferString(val))

	data, err := rd.ReadAll()
	if err != nil {
		return err
	}

	if len(data) != 1 {
		return fmt.Errorf("%w: permission scope contains invalid CSV record", ErrJWTValidation)
	}

	res := make(PermissionScopes, 0, len(data[0]))
	for _, scope := range data[0] {
		res = append(res, PermissionScope(scope))
	}

	if err := res.Validate(); err != nil {
		return err
	}

	*c = res

	return nil
}

func (c *Claims) Validate() error {
	var err error
	c.iss, err = sdk.AccAddressFromBech32(c.Issuer)
	if err != nil {
		return fmt.Errorf("%w: invalid issuer: %w", ErrJWTValidation, err)
	}

	// validate only the presence of the following fields.
	// JWT library is doing actual value validation
	if c.IssuedAt == nil {
		return fmt.Errorf("%w: missing issued at", ErrJWTValidation)
	}

	if c.NotBefore == nil {
		return fmt.Errorf("%w: missing not before", ErrJWTValidation)
	}

	if c.ExpiresAt == nil {
		return fmt.Errorf("%w: missing expiration time", ErrJWTValidation)
	}

	if c.Version != "v1" {
		return fmt.Errorf("%w: invalid version", ErrJWTValidation)
	}

	err = c.Leases.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (c *Leases) Validate() error {
	switch c.Access {
	case AccessTypeFull:
		if len(c.Permissions) > 0 {
			return fmt.Errorf("%w: full access cannot have permissions set", ErrJWTValidation)
		}

		if len(c.Scope) > 0 {
			return fmt.Errorf("%w: full access cannot have scope set", ErrJWTValidation)
		}
	case AccessTypeScoped:
		if c.Permissions != nil || len(c.Permissions) > 0 {
			return fmt.Errorf("%w: scoped access cannot have permissions set", ErrJWTValidation)
		}
		if len(c.Scope) == 0 {
			return fmt.Errorf("%w: scope access must have scope set", ErrJWTValidation)
		}

		if err := c.Scope.Validate(); err != nil {
			return err
		}
	case AccessTypeGranular:
		if len(c.Permissions) == 0 {
			return fmt.Errorf("%w: granular access must have permissions set", ErrJWTValidation)
		}

		err := c.Permissions.Validate()
		if err != nil {
			return fmt.Errorf("%w: invalid permissions: %w", ErrJWTValidation, err)
		}
	case "":
		return fmt.Errorf("%w: missing access type", ErrJWTValidation)
	default:
		return fmt.Errorf("%w: invalid access type '%s'", ErrJWTValidation, c.Access)
	}

	return nil
}

func (p Permissions) Validate() error {
	providers := make(map[string]bool)
	for _, permission := range p {
		// AccAddress.UnmarshalJSON does not validate an empty case
		if permission.Provider.Empty() {
			return fmt.Errorf("%w: empty provider", ErrJWTValidation)
		}

		if _, ok := providers[permission.Provider.String()]; ok {
			return fmt.Errorf("%w: duplicate provider %s", ErrJWTValidation, permission.Provider.String())
		}

		providers[permission.Provider.String()] = true

		switch permission.Access {
		case AccessTypeFull:
			if len(permission.Deployments) > 0 {
				return fmt.Errorf("%w: permission with full access cannot have deployments set", ErrJWTValidation)
			}
			if len(permission.Scope) > 0 {
				return fmt.Errorf("%w: permission with full access cannot have scope set", ErrJWTValidation)
			}
		case AccessTypeScoped:
			if permission.Deployments != nil || len(permission.Deployments) > 0 {
				return fmt.Errorf("%w: permission with scoped access cannot have deployments set", ErrJWTValidation)
			}
			if len(permission.Scope) == 0 {
				return fmt.Errorf("%w: permission with scope access must have scope set", ErrJWTValidation)
			}

			if err := permission.Scope.Validate(); err != nil {
				return err
			}
		case AccessTypeGranular:
			if permission.Scope != nil || len(permission.Scope) > 0 {
				return fmt.Errorf("%w: permission with granular access cannot have scope set", ErrJWTValidation)
			}

			if len(permission.Deployments) == 0 {
				return fmt.Errorf("%w: granular access must have permissions set", ErrJWTValidation)
			}

			for _, deployment := range permission.Deployments {
				if len(deployment.Scope) == 0 {
					return fmt.Errorf("%w: deployment permission scope must be set", ErrJWTValidation)
				}

				if deployment.DSeq == 0 {
					return fmt.Errorf("%w: dseq must be set", ErrJWTValidation)
				}

				if deployment.OSeq > 0 && deployment.GSeq == 0 {
					return fmt.Errorf("%w: gseq must be set if oseq is set", ErrJWTValidation)
				}

				if len(deployment.Scope) == 0 {
					return fmt.Errorf("%w: empty permissions scope", ErrJWTValidation)
				}

				if err := deployment.Scope.Validate(); err != nil {
					return err
				}

				if len(deployment.Services) == 0 {
					return fmt.Errorf("%w: empty services", ErrJWTValidation)
				}

				services := make(map[string]bool)

				for _, svc := range deployment.Services {
					if _, ok := services[svc]; ok {
						return fmt.Errorf("%w: duplicate service %s", ErrJWTValidation, svc)
					}

					services[svc] = true
				}
			}
		case "":
			return fmt.Errorf("%w: missing access type", ErrJWTValidation)
		default:
			return fmt.Errorf("%w: permission with invalid access type '%s'", ErrJWTValidation, permission.Access)
		}
	}

	return nil
}

func (c PermissionScopes) Validate() error {
	scopes := make(map[string]bool)

	for _, scope := range c {
		if _, ok := validScopes[scope]; !ok {
			return fmt.Errorf("%w: invalid scope %s", ErrJWTValidation, scope)
		}

		if _, ok := scopes[scope.String()]; ok {
			return fmt.Errorf("%w: duplicate scope %s", ErrJWTValidation, scope)
		}

		scopes[scope.String()] = true
	}

	return nil
}

func (c Claims) AuthorizeForProvider(acc sdk.Address) bool {
	switch c.Leases.Access {
	case AccessTypeFull:
		return true
	case AccessTypeScoped:
		return true
	case AccessTypeGranular:
		for _, perm := range c.Leases.Permissions {
			if perm.Provider.Equals(acc) {
				return true
			}
		}
	}

	return false
}

func (c Claims) AuthorizeForDeploymentID(did dtypes.DeploymentID, provider sdk.Address) bool {
	if c.Issuer != did.Owner {
		return false
	}

	switch c.Leases.Access {
	case AccessTypeFull:
		return true
	case AccessTypeScoped:
		return true
	case AccessTypeGranular:
		for _, perm := range c.Leases.Permissions {
			if perm.Provider.Equals(provider) {
				switch perm.Access {
				case AccessTypeFull:
					return true
				case AccessTypeScoped:
					return true
				case AccessTypeGranular:
					for _, depl := range perm.Deployments {
						if depl.DSeq == did.DSeq {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (c Claims) AuthorizeForLeaseID(lid mtypes.LeaseID) bool {
	if c.Issuer != lid.Owner {
		return false
	}

	switch c.Leases.Access {
	case AccessTypeFull:
		return true
	case AccessTypeScoped:
		return true
	case AccessTypeGranular:
		for _, perm := range c.Leases.Permissions {
			if perm.Provider.String() == lid.Provider {
				switch perm.Access {
				case AccessTypeFull:
					return true
				case AccessTypeScoped:
					return true
				case AccessTypeGranular:
					for _, depl := range perm.Deployments {
						if (depl.DSeq == lid.DSeq) &&
							(depl.GSeq == 0 || depl.GSeq == lid.GSeq) &&
							(depl.OSeq == 0 || depl.OSeq == lid.OSeq) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (c Claims) AuthorizeDeploymentIDForPermissionScope(did dtypes.DeploymentID, provider sdk.Address, scope PermissionScope) bool {
	if c.Issuer != did.Owner {
		return false
	}

	switch c.Leases.Access {
	case AccessTypeFull:
		return true
	case AccessTypeScoped:
		return c.Leases.Scope.Has(scope)
	case AccessTypeGranular:
		for _, perm := range c.Leases.Permissions {
			if perm.Provider.Equals(provider) {
				switch perm.Access {
				case AccessTypeFull:
					return true
				case AccessTypeScoped:
					return perm.Scope.Has(scope)
				case AccessTypeGranular:
					for _, depl := range perm.Deployments {
						if depl.DSeq == did.DSeq && depl.Scope.Has(scope) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (c Claims) AuthorizeLeaseIDForPermissionScope(lid mtypes.LeaseID, scope PermissionScope) bool {
	if c.Issuer != lid.Owner {
		return false
	}

	switch c.Leases.Access {
	case AccessTypeFull:
		return true
	case AccessTypeScoped:
		return c.Leases.Scope.Has(scope)
	case AccessTypeGranular:
		for _, perm := range c.Leases.Permissions {
			if perm.Provider.String() == lid.Provider {
				switch perm.Access {
				case AccessTypeFull:
					return true
				case AccessTypeScoped:
					return perm.Scope.Has(scope)
				case AccessTypeGranular:
					for _, depl := range perm.Deployments {
						if (depl.DSeq == lid.DSeq) &&
							(depl.GSeq == 0 || depl.GSeq == lid.GSeq) &&
							(depl.OSeq == 0 || depl.OSeq == lid.OSeq) &&
							depl.Scope.Has(scope) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (c Claims) IssuerAddress() sdk.Address {
	return c.iss
}
