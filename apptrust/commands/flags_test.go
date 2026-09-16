package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplicationCommandsIncludeAutoPromoteStagesFlag(t *testing.T) {
	for _, command := range []string{AppCreate, AppUpdate} {
		t.Run(command, func(t *testing.T) {
			var found bool
			for _, flag := range GetCommandFlags(command) {
				if flag.GetName() == AutoPromoteStagesFlag {
					found = true
					break
				}
			}
			assert.True(t, found, "%s should expose --%s", command, AutoPromoteStagesFlag)
		})
	}
}
