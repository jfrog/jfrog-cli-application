//go:build e2e

package e2e

import (
	"testing"

	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	output := utils.AppTrustCli.RunCliCmdWithOutput(t, "ping")
	assert.Contains(t, output, "OK")
}
