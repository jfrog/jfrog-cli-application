package version

import (
	"errors"
	"testing"

	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
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
				Modifications:  model.DistributionModifications{},
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
				DistributionRules: []model.DistributionRule{{}},
				Modifications: model.DistributionModifications{
					PathMappings: []model.DistributionPathMapping{
						{Input: "^my-repo/(.*)$", Output: "edge/$1"},
					},
				},
			},
		},
		{
			name: "no distribution flags produces a single empty rule",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsPayload: &model.DistributeAppVersionRequest{
				DistributionRules: []model.DistributionRule{{}},
				Modifications:     model.DistributionModifications{},
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
