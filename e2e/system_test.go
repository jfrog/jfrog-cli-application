//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	output := AppTrustCli.RunCliCmdWithOutput(t, "p")
	assert.Contains(t, output, "OK")
}
