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

func TestMonitorPolicyJSON(t *testing.T) {
	t.Run("minor patch sends minor_versions and patches_per_minor without value", func(t *testing.T) {
		minorVersions, patchesPerMinor := 2, 1
		payload, err := json.Marshal(MonitorPolicy{
			Type:            MonitorPolicyTypeVersionCountMinorPatch,
			MinorVersions:   &minorVersions,
			PatchesPerMinor: &patchesPerMinor,
		})
		require.NoError(t, err)

		assert.JSONEq(t, `{"type":"version_count_minor_patch","minor_versions":2,"patches_per_minor":1}`, string(payload))
	})

	t.Run("existing types send only type and value", func(t *testing.T) {
		value := 5
		payload, err := json.Marshal(MonitorPolicy{Type: MonitorPolicyTypeVersionCount, Value: &value})
		require.NoError(t, err)

		assert.JSONEq(t, `{"type":"version_count","value":5}`, string(payload))
	})
}
