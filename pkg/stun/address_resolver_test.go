package stun

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStunAddressResolver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sw, err := NewStunAddressResolver(ctx, "stun.l.google.com:19302")
	require.NoError(t, err)
	require.NotNil(t, sw)

	err = sw.Resolve()
	require.NoError(t, err)
	require.NotEmpty(t, sw.publicAddr)

	require.NotEmpty(t, sw.PublicIP())
	require.NotEmpty(t, sw.PublicPort())

	require.NoError(t, sw.conn.Close())

	t.Logf("Public IP: %s", sw.PublicIP())
	t.Logf("Public Port: %d", sw.PublicPort())
}
