package stun

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewChecker(t *testing.T) {
	ctx := context.Background()
	stunServer := "stun.l.google.com:19302"
	checker := NewChecker(ctx, stunServer, false)
	require.NotNil(t, checker)

	require.Equal(t, 0, len(checker.observers))
	require.Equal(t, NatBehavior_NAT_BEHAVIOR_UNKNOWN, checker.mappingBehavior)
	require.Equal(t, NatBehavior_NAT_BEHAVIOR_UNKNOWN, checker.filteringBehavior)
	require.Equal(t, NatType_NAT_TYPE_UNKNOWN, checker.natTpye)

	as, err := checker.NewAddress()
	require.NoError(t, err)
	require.NotNil(t, as)

	require.NotEmpty(t, as.PublicAddr())
	require.NotEmpty(t, as.PublicPort())
}

type testObserver struct {
	t *testing.T
}

func (o *testObserver) OnMappingBehaviorChanged(behavior NatBehavior) {
	o.t.Logf("mapping behavior changed: %v", behavior)
}

func (o *testObserver) OnFilteringBehaviorChanged(behavior NatBehavior) {
	o.t.Logf("filtering behavior changed: %v", behavior)
}

func (o *testObserver) OnNatTypeChanged(natType NatType) {
	o.t.Logf("nat type changed: %v", natType)
}

func TestNewChecker2(t *testing.T) {
	ctx := context.Background()
	// stunServer := "stun.syncthing.net:3478"
	stunServer := "114.55.134.60:3478"
	enableNatCheck := true
	checker := NewChecker(ctx, stunServer, enableNatCheck, &testObserver{t: t})
	require.NotNil(t, checker)

	require.Equal(t, 1, len(checker.observers))

	time.Sleep(20 * time.Second)

	as, err := checker.NewAddress()
	require.NoError(t, err)
	require.NotNil(t, as)

	require.NotEmpty(t, as.PublicAddr())
	require.NotEmpty(t, as.PublicPort())
	as.conn.Close()
}
