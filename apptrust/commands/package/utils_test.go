package packagecmds

import (
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/stretchr/testify/assert"
)

func TestBuildPackageRequestPayload(t *testing.T) {
	tests := []struct {
		name             string
		arguments        []string
		expectedKey      string
		expectedType     string
		expectedName     string
		expectedVersions []string
	}{
		{
			name:             "Valid package request with single version",
			arguments:        []string{"my-app", "npm", "my-package", "1.0.0"},
			expectedKey:      "my-app",
			expectedType:     "npm",
			expectedName:     "my-package",
			expectedVersions: []string{"1.0.0"},
		},
		{
			name:             "Valid package request with multiple versions",
			arguments:        []string{"my-app", "npm", "my-package", "1.0.0,1.1.0 , 1.2.0"},
			expectedKey:      "my-app",
			expectedType:     "npm",
			expectedName:     "my-package",
			expectedVersions: []string{"1.0.0", "1.1.0", "1.2.0"},
		},
		{
			name:             "Package request without versions",
			arguments:        []string{"my-app", "npm", "my-package"},
			expectedKey:      "my-app",
			expectedType:     "npm",
			expectedName:     "my-package",
			expectedVersions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{
				Arguments: tt.arguments,
			}

			result, err := BuildPackageRequestPayload(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedKey, result.ApplicationKey)
			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedVersions, result.Versions)
		})
	}
}
