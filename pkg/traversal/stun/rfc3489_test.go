package stun

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNATType(t *testing.T) {
	nt, err := NATType("stun.syncthing.net:3478")
	require.NoError(t, err)
	require.NotEmpty(t, nt)

	t.Logf("NAT type: %v", nt)
}
