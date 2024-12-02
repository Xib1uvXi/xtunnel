package stun

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMappingTests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := MappingTests(ctx, "stun.syncthing.net:3478")
	require.NoError(t, err)

	require.NotEmpty(t, result)

	t.Logf("result: %v", result)
}

func TestFilteringTests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := FilteringTests(ctx, "stun.syncthing.net:3478")
	require.NoError(t, err)

	require.NotEmpty(t, result)

	t.Logf("result: %v", result)
}
