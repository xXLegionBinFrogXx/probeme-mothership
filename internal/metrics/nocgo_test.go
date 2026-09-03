//go:build !cgo

package metrics

import "testing"

func TestPackageBuildsWithoutCgo(t *testing.T) {
	t.Log("internal/metrics builds with CGO_ENABLED=0")
}
