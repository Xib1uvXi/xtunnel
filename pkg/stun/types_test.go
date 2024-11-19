package stun

import (
	"testing"
)

func TestNatTypeString(t *testing.T) {
	tests := []struct {
		natType  NatType
		expected string
	}{
		{NatType_NAT_TYPE_UNKNOWN, "NAT_TYPE_UNKNOWN"},
		{NatType_NAT_TYPE_FULL_CONE, "NAT_TYPE_FULL_CONE"},
		{NatType_NAT_TYPE_RESTRICTED_CONE, "NAT_TYPE_RESTRICTED_CONE"},
		{NatType_NAT_TYPE_PORT_RESTRICTED_CONE, "NAT_TYPE_PORT_RESTRICTED_CONE"},
		{NatType_NAT_TYPE_SYMMETRIC, "NAT_TYPE_SYMMETRIC"},
		{NatType_NAT_TYPE_NONE, "NAT_TYPE_NONE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.natType.String(); got != tt.expected {
				t.Errorf("NatType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestNatBehaviorString(t *testing.T) {
	tests := []struct {
		natBehavior NatBehavior
		expected    string
	}{
		{NatBehavior_NAT_BEHAVIOR_UNKNOWN, "NAT_BEHAVIOR_UNKNOWN"},
		{NatBehavior_NAT_BEHAVIOR_ENDPOINT_INDEPENDENT, "NAT_BEHAVIOR_ENDPOINT_INDEPENDENT"},
		{NatBehavior_NAT_BEHAVIOR_ADDRESS_DEPENDENT, "NAT_BEHAVIOR_ADDRESS_DEPENDENT"},
		{NatBehavior_NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT, "NAT_BEHAVIOR_ADDRESS_AND_PORT_DEPENDENT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.natBehavior.String(); got != tt.expected {
				t.Errorf("NatBehavior.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
