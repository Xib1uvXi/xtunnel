package xkcp

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetBlockCrypt(t *testing.T) {
	bc := GetBlockCrypt("testseed")
	require.NotNil(t, bc)
}
