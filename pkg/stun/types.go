package stun

type NatType int32

// String returns the string representation of the NatType.
func (n NatType) String() string {
	return NatType_name[int32(n)]
}

const (
	// Unknown
	NatType_NAT_TYPE_UNKNOWN NatType = 0
	// Full Cone
	NatType_NAT_TYPE_FULL_CONE NatType = 1
	// Restricted Cone
	NatType_NAT_TYPE_RESTRICTED_CONE NatType = 2
	// Port Restricted Cone
	NatType_NAT_TYPE_PORT_RESTRICTED_CONE NatType = 3
	// Symmetric
	NatType_NAT_TYPE_SYMMETRIC NatType = 4
	// None
	NatType_NAT_TYPE_NONE NatType = 5
)

// Enum value maps for NatType.
var (
	NatType_name = map[int32]string{
		0: "NAT_TYPE_UNKNOWN",
		1: "NAT_TYPE_FULL_CONE",
		2: "NAT_TYPE_RESTRICTED_CONE",
		3: "NAT_TYPE_PORT_RESTRICTED_CONE",
		4: "NAT_TYPE_SYMMETRIC",
		5: "NAT_TYPE_NONE",
	}
	NatType_value = map[string]int32{
		"NAT_TYPE_UNKNOWN":              0,
		"NAT_TYPE_FULL_CONE":            1,
		"NAT_TYPE_RESTRICTED_CONE":      2,
		"NAT_TYPE_PORT_RESTRICTED_CONE": 3,
		"NAT_TYPE_SYMMETRIC":            4,
		"NAT_TYPE_NONE":                 5,
	}
)

type NatBehavior int32

// String returns the string representation of the NatBehavior.
func (n NatBehavior) String() string {
	return NatBehavior_name[int32(n)]
}

const (
	// Unknown
	NatBehavior_NAT_BEHAVIOR_UNKNOWN NatBehavior = 0
	// Endpoint Independent
	NatBehavior_NAT_BEHAVIOR_ENDPOINT_INDEPENDENT NatBehavior = 1
	// Address Dependent
	NatBehavior_NAT_BEHAVIOR_ADDRESS_DEPENDENT NatBehavior = 2
	// Address Port Dependent
	NatBehavior_NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT NatBehavior = 3
)

// Enum value maps for NatBehavior.
var (
	NatBehavior_name = map[int32]string{
		0: "NAT_BEHAVIOR_UNKNOWN",
		1: "NAT_BEHAVIOR_ENDPOINT_INDEPENDENT",
		2: "NAT_BEHAVIOR_ADDRESS_DEPENDENT",
		3: "NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT",
	}
	NatBehavior_value = map[string]int32{
		"NAT_BEHAVIOR_UNKNOWN":                    0,
		"NAT_BEHAVIOR_ENDPOINT_INDEPENDENT":       1,
		"NAT_BEHAVIOR_ADDRESS_DEPENDENT":          2,
		"NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT": 3,
	}
)
