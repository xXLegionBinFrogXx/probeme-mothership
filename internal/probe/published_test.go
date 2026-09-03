package probe

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xXLegionBinFrogXx/probeme-mothership/internal/provider"
)

func TestPublishedLatestWins(t *testing.T) {
	pub := NewPublished()
	require.False(t, pub.Ready())

	first := &provider.Snapshot{Valid: 0}
	pub.Store(first)
	require.Same(t, first, pub.Load())
	require.False(t, pub.Ready())

	second := &provider.Snapshot{Valid: provider.CapCPU}
	pub.Store(second)
	require.Same(t, second, pub.Load())
	require.True(t, pub.Ready())
}
