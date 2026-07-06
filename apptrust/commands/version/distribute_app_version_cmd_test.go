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

func TestDistributeAppVersionCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"
	requestPayload := &model.DistributeAppVersionRequest{
		DistributionRules: []model.DistributionRule{{SiteName: "edge-*"}},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().DistributeAppVersion(gomock.Any(), applicationKey, version, requestPayload).
		Return(nil).Times(1)

	cmd := &distributeAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestDistributeAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"
	requestPayload := &model.DistributeAppVersionRequest{
		DistributionRules: []model.DistributionRule{{SiteName: "edge-*"}},
	}
	expectedError := errors.New("service error occurred")

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().DistributeAppVersion(gomock.Any(), applicationKey, version, requestPayload).
		Return(expectedError).Times(1)

	cmd := &distributeAppVersionCommand{
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

func TestDistributeAppVersionCommand_FlagsSuite(t *testing.T) {
	tests := []struct {
		name           string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsPayload *model.DistributeAppVersionRequest
	}{
		{
			name: "site, city and country codes with create-repo",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SiteFlag, "edge-*")
				ctx.AddStringFlag(commands.CityFlag, "NYC")
				ctx.AddStringFlag(commands.CountryCodesFlag, "US;CA")
				ctx.AddBoolFlag(commands.CreateRepoFlag, true)
			},
			expectsPayload: &model.DistributeAppVersionRequest{
				DistributionRules: []model.DistributionRule{
					{SiteName: "edge-*", CityName: "NYC", CountryCodes: []string{"US", "CA"}},
				},
				AutoCreateRepo: true,
			},
		},
		{
			name: "mapping pattern and target",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.MappingPatternFlag, "my-repo/(*)")
				ctx.AddStringFlag(commands.MappingTargetFlag, "edge/{1}")
			},
			expectsPayload: &model.DistributeAppVersionRequest{
				DistributionRules: []model.DistributionRule{{SiteName: "*"}},
				Modifications: &model.DistributionModifications{
					PathMappings: []model.DistributionPathMapping{
						{Input: "^my-repo/(.*)$", Output: "edge/$1"},
					},
				},
			},
		},
		{
			name: "no distribution flags defaults to distributing to all targets",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsPayload: &model.DistributeAppVersionRequest{
				DistributionRules: []model.DistributionRule{{SiteName: "*"}},
			},
		},
		{
			name: "only mapping-pattern provided",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.MappingPatternFlag, "my-repo/(*)")
			},
			expectsError:  true,
			errorContains: "must be provided together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			tt.ctxSetup(ctx)
			ctx.AddStringFlag("url", "https://example.com")

			var actualPayload *model.DistributeAppVersionRequest
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().DistributeAppVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, _ string, req *model.DistributeAppVersionRequest) error {
						actualPayload = req
						return nil
					}).Times(1)
			}

			cmd := &distributeAppVersionCommand{
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

func TestDistributeAppVersionCommand_SpecAndFlags(t *testing.T) {
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

			cmd := &distributeAppVersionCommand{
				versionService: mockversions.NewMockVersionService(ctrl),
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "can't be used with")
		})
	}
}

func TestDistributeAppVersionCommand_SpecFileSuite(t *testing.T) {
	t.Run("payload built from dist-rules file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		content := `{"distribution_rules":[{"site_name":"site-1","city_name":"city-1","country_codes":["US"]},{"site_name":"site-2"}]}`
		filePath := filepath.Join(t.TempDir(), "dist-rules.json")
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		ctx := &components.Context{}
		ctx.Arguments = []string{"app-key", "1.0.0"}
		ctx.AddStringFlag(commands.DistRulesFlag, filePath)
		ctx.AddStringFlag("url", "https://example.com")

		var actualPayload *model.DistributeAppVersionRequest
		mockVersionService := mockversions.NewMockVersionService(ctrl)
		mockVersionService.EXPECT().DistributeAppVersion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ interface{}, _ string, _ string, req *model.DistributeAppVersionRequest) error {
				actualPayload = req
				return nil
			}).Times(1)

		cmd := &distributeAppVersionCommand{
			versionService: mockVersionService,
		}

		err := cmd.prepareAndRunCommand(ctx)
		require.NoError(t, err)
		assert.Equal(t, &model.DistributeAppVersionRequest{
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
		ctx.AddStringFlag("url", "https://example.com")

		cmd := &distributeAppVersionCommand{
			versionService: mockversions.NewMockVersionService(ctrl),
		}

		err := cmd.prepareAndRunCommand(ctx)
		assert.Error(t, err)
	})
}

func TestDistributeAppVersionCommand_WrongNumberOfArguments(t *testing.T) {
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

			cmd := &distributeAppVersionCommand{
				versionService: mockversions.NewMockVersionService(ctrl),
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.Error(t, err)
		})
	}
}
