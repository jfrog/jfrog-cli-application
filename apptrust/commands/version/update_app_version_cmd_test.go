package version

import (
	"errors"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestUpdateAppVersionCommand_Run(t *testing.T) {
	tests := []struct {
		name         string
		request      *model.UpdateAppVersionRequest
		shouldError  bool
		errorMessage string
	}{
		{
			name: "success",
			request: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "release/1.2.3",
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
			},
		},
		{
			name: "context error",
			request: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "test-tag",
			},
			shouldError:  true,
			errorMessage: "context error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if tt.shouldError {
				mockVersionService.EXPECT().UpdateAppVersion(gomock.Any(), gomock.Any()).
					Return(errors.New(tt.errorMessage)).Times(1)
			} else {
				mockVersionService.EXPECT().UpdateAppVersion(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			}

			cmd := &updateAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  &config.ServerDetails{Url: "https://example.com"},
				applicationKey: "app-key",
				version:        "1.0.0",
				requestPayload: tt.request,
			}

			err := cmd.Run()
			if tt.shouldError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAppVersionCommand_FlagsSuite(t *testing.T) {
	tests := []struct {
		name           string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsPayload *model.UpdateAppVersionRequest
	}{
		{
			name: "tag only",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.TagFlag, "release/1.2.3")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "release/1.2.3",
			},
		},
		{
			name: "properties only - single value",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.PropertiesFlag, "status=rc")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Properties: map[string][]string{
					"status": {"rc"},
				},
			},
		},
		{
			name: "properties only - multiple values",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.PropertiesFlag, "status=rc,validated")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
			},
		},
		{
			name: "properties only - multiple properties",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.PropertiesFlag, "status=rc,validated;deployed_to=staging-A,staging-B")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Properties: map[string][]string{
					"status":      {"rc", "validated"},
					"deployed_to": {"staging-A", "staging-B"},
				},
			},
		},
		{
			name: "delete properties only",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.DeletePropertyFlag, "legacy_param;toBeDeleted")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey:   "app-key",
				Version:          "1.0.0",
				DeleteProperties: []string{"legacy_param", "toBeDeleted"},
			},
		},
		{
			name: "empty properties (clears values)",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.PropertiesFlag, "old_feature_flag=")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Properties: map[string][]string{
					"old_feature_flag": {},
				},
			},
		},
		{
			name: "combined update",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.TagFlag, "release/1.2.3")
				ctx.AddStringFlag(commands.PropertiesFlag, "status=rc,validated")
				ctx.AddStringFlag(commands.DeletePropertyFlag, "old_param")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "release/1.2.3",
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
				DeleteProperties: []string{"old_param"},
			},
		},
		{
			name: "empty tag (removes tag)",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.TagFlag, "")
			},
			expectsPayload: &model.UpdateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "",
			},
		},
		{
			name: "invalid property format",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.PropertiesFlag, "invalid-format")
			},
			expectsError:  true,
			errorContains: "invalid property format",
		},
		{
			name: "empty property key",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.PropertiesFlag, "=value")
			},
			expectsError:  true,
			errorContains: "property key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			tt.ctxSetup(ctx)
			ctx.AddStringFlag("url", "https://example.com")

			var actualPayload *model.UpdateAppVersionRequest
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().UpdateAppVersion(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.UpdateAppVersionRequest) error {
						actualPayload = req
						return nil
					}).Times(1)
			}

			cmd := &updateAppVersionCommand{
				versionService: mockVersionService,
			}

			err := cmd.prepareAndRunCommand(ctx)
			if tt.expectsError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectsPayload, actualPayload)
			}
		})
	}
}

func TestParseProperties(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string][]string
		expectErr bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:  "single property with single value",
			input: "status=rc",
			expected: map[string][]string{
				"status": {"rc"},
			},
		},
		{
			name:  "single property with multiple values",
			input: "status=rc,validated",
			expected: map[string][]string{
				"status": {"rc", "validated"},
			},
		},
		{
			name:  "multiple properties",
			input: "status=rc,validated;deployed_to=staging-A,staging-B",
			expected: map[string][]string{
				"status":      {"rc", "validated"},
				"deployed_to": {"staging-A", "staging-B"},
			},
		},
		{
			name:  "empty values (clears values)",
			input: "old_feature_flag=",
			expected: map[string][]string{
				"old_feature_flag": {},
			},
		},
		{
			name:  "with spaces",
			input: " status = rc , validated ; deployed_to = staging-A , staging-B ",
			expected: map[string][]string{
				"status":      {"rc", "validated"},
				"deployed_to": {"staging-A", "staging-B"},
			},
		},
		{
			name:      "invalid format - missing =",
			input:     "invalid-format",
			expectErr: true,
		},
		{
			name:      "empty key",
			input:     "=value",
			expectErr: true,
		},
		{
			name:      "empty key with spaces",
			input:     " =value",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ParseListPropertiesFlag(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
