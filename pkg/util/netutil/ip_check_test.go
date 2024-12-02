package netutil

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsSameSegment(t *testing.T) {
	ip1 := "192.168.2.1"
	ip2 := "192.30.2.1"

	require.False(t, IsSameSegmentIP(ip1, ip2, 2))

	ip1 = "192.168.2.1"
	ip2 = "192.168.3.1"

	require.True(t, IsSameSegmentIP(ip1, ip2, 2))

	ip1 = "192.168.3.1"
	ip2 = "192.168.3.2"

	require.True(t, IsSameSegmentIP(ip1, ip2, 3))

	ip1 = "192.168.3.1"
	ip2 = "192.168.2.2"

	require.False(t, IsSameSegmentIP(ip1, ip2, 3))

	ip1 = "192.168.2.1"
	ip2 = "192.168.2.1"

	require.True(t, IsSameSegmentIP(ip1, ip2, 4))
}
