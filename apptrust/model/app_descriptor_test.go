package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppDescriptorAutoPromoteStagesJSON(t *testing.T) {
	t.Run("omitted when flag is not supplied", func(t *testing.T) {
		payload, err := json.Marshal(AppDescriptor{ApplicationKey: "app-key"})
		require.NoError(t, err)

		assert.NotContains(t, string(payload), "auto_promote_stages")
	})

	t.Run("preserves configured order", func(t *testing.T) {
		stages := []string{"DEV", "PROD"}
		payload, err := json.Marshal(AppDescriptor{
			ApplicationKey:    "app-key",
			AutoPromoteStages: &stages,
		})
		require.NoError(t, err)

		assert.JSONEq(t, `{"application_key":"app-key","auto_promote_stages":["DEV","PROD"]}`, string(payload))
	})

	t.Run("includes empty list to disable auto-promotion", func(t *testing.T) {
		stages := []string{}
		payload, err := json.Marshal(AppDescriptor{
			ApplicationKey:    "app-key",
			AutoPromoteStages: &stages,
		})
		require.NoError(t, err)

		assert.JSONEq(t, `{"application_key":"app-key","auto_promote_stages":[]}`, string(payload))
	})
}
