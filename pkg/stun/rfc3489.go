package stun

import (
	"fmt"

	gostun "github.com/ccding/go-stun/stun"
)

// implements the STUN RFC 5780
// https://tools.ietf.org/html/rfc3489
// from github.com/ccding/go-stun/stun

func NATType(serverAddr string) (NatType, error) {
	client := gostun.NewClient()
	client.SetServerAddr(serverAddr)

	nat, _, err := client.Discover()
	if err != nil {
		return NatType_NAT_TYPE_UNKNOWN, err
	}

	defer client.Keepalive()

	return initNATType(nat)
}

func initNATType(sNatType gostun.NATType) (NatType, error) {
	switch sNatType {
	case gostun.NATNone:
		return NatType_NAT_TYPE_NONE, nil
	case gostun.NATFull:
		return NatType_NAT_TYPE_FULL_CONE, nil
	case gostun.NATRestricted:
		return NatType_NAT_TYPE_RESTRICTED_CONE, nil
	case gostun.NATPortRestricted:
		return NatType_NAT_TYPE_PORT_RESTRICTED_CONE, nil
	case gostun.NATSymetric:
		return NatType_NAT_TYPE_SYMMETRIC, nil
	default:
		return NatType_NAT_TYPE_UNKNOWN, fmt.Errorf("unsupported NAT type: %v", sNatType)
	}
}
