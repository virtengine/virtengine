package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "roles"

	// StoreKey is the store key string for roles module
	StoreKey = ModuleName

	// RouterKey is the message route for roles module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for roles module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixRoleAssignment is the prefix for role assignment storage
	// Key: PrefixRoleAssignment | address | role -> RoleAssignment
	PrefixRoleAssignment = []byte{0x01}

	// PrefixAccountState is the prefix for account state storage
	// Key: PrefixAccountState | address -> AccountStateRecord
	PrefixAccountState = []byte{0x02}

	// PrefixRoleMembers is the prefix for role members index
	// Key: PrefixRoleMembers | role | address -> []byte{1}
	PrefixRoleMembers = []byte{0x03}

	// PrefixGenesisAccount is the prefix for genesis account addresses
	// Key: PrefixGenesisAccount | address -> []byte{1}
	PrefixGenesisAccount = []byte{0x04}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x05}

	// PrefixSanction is the prefix for sanction records
	// Key: PrefixSanction | sanction ID -> Sanction
	PrefixSanction = []byte{0x06}

	// PrefixSubjectSanction is the prefix for the subject -> sanction index
	// Key: PrefixSubjectSanction | subject | sanction ID -> []byte{1}
	PrefixSubjectSanction = []byte{0x07}

	// PrefixSanctionSequence is the prefix for the sanction ID sequence
	// Key: PrefixSanctionSequence -> big-endian uint64
	PrefixSanctionSequence = []byte{0x08}
)

// RoleAssignmentKey returns the store key for a role assignment
func RoleAssignmentKey(address []byte, role Role) []byte {
	key := make([]byte, 0, len(PrefixRoleAssignment)+len(address)+1)
	key = append(key, PrefixRoleAssignment...)
	key = append(key, address...)
	key = append(key, byte(role))
	return key
}

// RoleAssignmentPrefixKey returns the store key prefix for all roles of an address
func RoleAssignmentPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixRoleAssignment)+len(address))
	key = append(key, PrefixRoleAssignment...)
	key = append(key, address...)
	return key
}

// AccountStateKey returns the store key for an account state
func AccountStateKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixAccountState)+len(address))
	key = append(key, PrefixAccountState...)
	key = append(key, address...)
	return key
}

// RoleMembersKey returns the store key for a role member
func RoleMembersKey(role Role, address []byte) []byte {
	key := make([]byte, 0, len(PrefixRoleMembers)+1+len(address))
	key = append(key, PrefixRoleMembers...)
	key = append(key, byte(role))
	key = append(key, address...)
	return key
}

// RoleMembersPrefixKey returns the store key prefix for all members of a role
func RoleMembersPrefixKey(role Role) []byte {
	key := make([]byte, 0, len(PrefixRoleMembers)+1)
	key = append(key, PrefixRoleMembers...)
	key = append(key, byte(role))
	return key
}

// GenesisAccountKey returns the store key for a genesis account
func GenesisAccountKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixGenesisAccount)+len(address))
	key = append(key, PrefixGenesisAccount...)
	key = append(key, address...)
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// SanctionKey returns the store key for a sanction record
func SanctionKey(sanctionID string) []byte {
	key := make([]byte, 0, len(PrefixSanction)+len(sanctionID))
	key = append(key, PrefixSanction...)
	key = append(key, []byte(sanctionID)...)
	return key
}

// SubjectSanctionPrefixKey returns the store key prefix for all sanction
// records belonging to a subject
func SubjectSanctionPrefixKey(subject string) []byte {
	key := make([]byte, 0, len(PrefixSubjectSanction)+len(subject)+1)
	key = append(key, PrefixSubjectSanction...)
	key = append(key, []byte(subject)...)
	key = append(key, '/')
	return key
}

// SubjectSanctionKey returns the index key linking a subject to a sanction
func SubjectSanctionKey(subject, sanctionID string) []byte {
	key := SubjectSanctionPrefixKey(subject)
	key = append(key, []byte(sanctionID)...)
	return key
}
