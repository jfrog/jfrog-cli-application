package version

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteDeleteAppVersionCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"
	requestPayload := &model.RemoteDeleteAppVersionRequest{
		DistributionRules: []model.DistributionRule{{SiteName: "edge-*"}},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().RemoteDeleteAppVersion(gomock.Any(), applicationKey, version, requestPayload).
		Return(nil).Times(1)

	cmd := &remoteDeleteAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestRemoteDeleteAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"
	requestPayload := &model.RemoteDeleteAppVersionRequest{
		DistributionRules: []model.DistributionRule{{SiteName: "edge-*"}},
	}
	expectedError := errors.New("service error occurred")

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().RemoteDeleteAppVersion(gomock.Any(), applicationKey, version, requestPayload).
		Return(expectedError).Times(1)

	cmd := &remoteDeleteAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error occurred")
}

func TestRemoteDeleteAppVersionCommand_FlagsSuite(t *testing.T) {
	tests := []struct {
		name           string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsPayload *model.RemoteDeleteAppVersionRequest
	}{
		{
			name: "site, city and country codes",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SiteFlag, "edge-*")
				ctx.AddStringFlag(commands.CityFlag, "NYC")
				ctx.AddStringFlag(commands.CountryCodesFlag, "US;CA")
			},
			expectsPayload: &model.RemoteDeleteAppVersionRequest{
				DistributionRules: []model.DistributionRule{
					{SiteName: "edge-*", CityName: "NYC", CountryCodes: []string{"US", "CA"}},
				},
			},
		},
		{
			name: "dry-run flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddBoolFlag(commands.DryRunFlag, true)
			},
			expectsPayload: &model.RemoteDeleteAppVersionRequest{
				DryRun:            true,
				DistributionRules: []model.DistributionRule{{SiteName: "*"}},
			},
		},
		{
			name: "no distribution flags defaults to distributing to all targets",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsPayload: &model.RemoteDeleteAppVersionRequest{
				DistributionRules: []model.DistributionRule{{SiteName: "*"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			tt.ctxSetup(ctx)

			// Skip the interactive confirmation prompt
			ctx.AddBoolFlag(commands.QuietFlag, true)

			ctx.AddStringFlag("url", "https://example.com")

			var actualPayload *model.RemoteDeleteAppVersionRequest
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().RemoteDeleteAppVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, _ string, req *model.RemoteDeleteAppVersionRequest) error {
						actualPayload = req
						return nil
					}).Times(1)
			}

			cmd := &remoteDeleteAppVersionCommand{
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

func TestRemoteDeleteAppVersionCommand_QuietSuite(t *testing.T) {
	tests := []struct {
		name        string
		quiet       bool
		expectsCall bool
	}{
		{name: "quiet skips confirmation and calls service", quiet: true, expectsCall: true},
		{name: "without quiet the default confirmation aborts", quiet: false, expectsCall: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			ctx.Arguments = []string{"app-key", "1.0.0"}
			ctx.AddStringFlag("url", "https://example.com")
			if tt.quiet {
				ctx.AddBoolFlag(commands.QuietFlag, true)
			}

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if tt.expectsCall {
				mockVersionService.EXPECT().RemoteDeleteAppVersion(gomock.Any(), "app-key", "1.0.0", gomock.Any()).
					Return(nil).Times(1)
			}

			cmd := &remoteDeleteAppVersionCommand{
				versionService: mockVersionService,
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.NoError(t, err)
		})
	}
}

func TestRemoteDeleteAppVersionCommand_SpecAndFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{name: "dist-rules with site", flagName: commands.SiteFlag},
		{name: "dist-rules with city", flagName: commands.CityFlag},
		{name: "dist-rules with country-codes", flagName: commands.CountryCodesFlag},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			ctx.Arguments = []string{"app-key", "1.0.0"}
			ctx.AddStringFlag(commands.DistRulesFlag, "rules.json")
			ctx.AddStringFlag(tt.flagName, "value")
			ctx.AddStringFlag("url", "https://example.com")

			cmd := &remoteDeleteAppVersionCommand{
				versionService: mockversions.NewMockVersionService(ctrl),
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "can't be used with")
		})
	}
}

func TestRemoteDeleteAppVersionCommand_SpecFileSuite(t *testing.T) {
	t.Run("payload built from dist-rules file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		content := `{"distribution_rules":[{"site_name":"site-1","city_name":"city-1","country_codes":["US"]},{"site_name":"site-2"}]}`
		filePath := filepath.Join(t.TempDir(), "dist-rules.json")
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		ctx := &components.Context{}
		ctx.Arguments = []string{"app-key", "1.0.0"}
		ctx.AddStringFlag(commands.DistRulesFlag, filePath)
		ctx.AddBoolFlag(commands.QuietFlag, true)
		ctx.AddStringFlag("url", "https://example.com")

		var actualPayload *model.RemoteDeleteAppVersionRequest
		mockVersionService := mockversions.NewMockVersionService(ctrl)
		mockVersionService.EXPECT().RemoteDeleteAppVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ interface{}, _ string, _ string, req *model.RemoteDeleteAppVersionRequest) error {
				actualPayload = req
				return nil
			}).Times(1)

		cmd := &remoteDeleteAppVersionCommand{
			versionService: mockVersionService,
		}

		err := cmd.prepareAndRunCommand(ctx)
		require.NoError(t, err)
		assert.Equal(t, &model.RemoteDeleteAppVersionRequest{
			DistributionRules: []model.DistributionRule{
				{SiteName: "site-1", CityName: "city-1", CountryCodes: []string{"US"}},
				{SiteName: "site-2"},
			},
		}, actualPayload)
	})

	t.Run("missing dist-rules file returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := &components.Context{}
		ctx.Arguments = []string{"app-key", "1.0.0"}
		ctx.AddStringFlag(commands.DistRulesFlag, filepath.Join(t.TempDir(), "does-not-exist.json"))
		ctx.AddBoolFlag(commands.QuietFlag, true)
		ctx.AddStringFlag("url", "https://example.com")

		cmd := &remoteDeleteAppVersionCommand{
			versionService: mockversions.NewMockVersionService(ctrl),
		}

		err := cmd.prepareAndRunCommand(ctx)
		assert.Error(t, err)
	})
}

func TestRemoteDeleteAppVersionCommand_DistributionRulesEmpty(t *testing.T) {
	tests := []struct {
		name  string
		rules []model.DistributionRule
		empty bool
	}{
		{
			name:  "nil rules",
			rules: nil,
			empty: true,
		},
		{
			name:  "no rules",
			rules: []model.DistributionRule{},
			empty: true,
		},
		{
			name:  "single empty rule",
			rules: []model.DistributionRule{{}},
			empty: true,
		},
		{
			name:  "single rule with site",
			rules: []model.DistributionRule{{SiteName: "edge-*"}},
			empty: false,
		},
		{
			name:  "single rule with city",
			rules: []model.DistributionRule{{CityName: "NYC"}},
			empty: false,
		},
		{
			name:  "single rule with country codes",
			rules: []model.DistributionRule{{CountryCodes: []string{"US"}}},
			empty: false,
		},
		{
			name:  "multiple rules",
			rules: []model.DistributionRule{{SiteName: "site-1"}, {SiteName: "site-2"}},
			empty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &remoteDeleteAppVersionCommand{
				requestPayload: &model.RemoteDeleteAppVersionRequest{DistributionRules: tt.rules},
			}
			assert.Equal(t, tt.empty, cmd.distributionRulesEmpty())
		})
	}
}

func TestRemoteDeleteAppVersionCommand_WrongNumberOfArguments(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
	}{
		{name: "no arguments", arguments: []string{}},
		{name: "single argument", arguments: []string{"app-key"}},
		{name: "too many arguments", arguments: []string{"app-key", "1.0.0", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				PrintCommandHelp: func(string) error { return nil },
			}
			ctx.Arguments = tt.arguments

			cmd := &remoteDeleteAppVersionCommand{
				versionService: mockversions.NewMockVersionService(ctrl),
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.Error(t, err)
		})
	}
}
