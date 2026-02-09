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

func TestUpdateAppVersionSourcesCommand_Run(t *testing.T) {
	tests := []struct {
		name         string
		request      *model.UpdateVersionSourcesRequest
		shouldError  bool
		errorMessage string
	}{
		{
			name: "success",
			request: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "pkg", Version: "1.0.0", Repository: "npm-local"},
					},
				},
			},
		},
		{
			name: "service error",
			request: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Builds: []model.CreateVersionBuild{
						{Name: "build1", Number: "100"},
					},
				},
			},
			shouldError:  true,
			errorMessage: "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if tt.shouldError {
				mockVersionService.EXPECT().UpdateAppVersionSources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New(tt.errorMessage)).Times(1)
			} else {
				mockVersionService.EXPECT().UpdateAppVersionSources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			}

			cmd := &updateAppVersionSourcesCommand{
				versionService: mockVersionService,
				serverDetails:  &config.ServerDetails{Url: "https://example.com"},
				applicationKey: "app-key",
				version:        "1.0.0",
				requestPayload: tt.request,
				sync:           true,
				dryRun:         false,
				failFast:       true,
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

func TestUpdateAppVersionSourcesCommand_SourceFlagsSuite(t *testing.T) {
	tests := []struct {
		name            string
		ctxSetup        func(*components.Context)
		expectsError    bool
		errorContains   string
		expectsPayload  *model.UpdateVersionSourcesRequest
		expectsSync     bool
		expectsDryRun   bool
		expectsFailFast bool
	}{
		{
			name: "update with source-type-packages and default flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=pkg,version=2.0.0,repo-key=npm-local")
			},
			expectsPayload: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "pkg", Version: "2.0.0", Repository: "npm-local"},
					},
				},
			},
			expectsSync:     true,
			expectsDryRun:   false,
			expectsFailFast: true,
		},
		{
			name: "update with source-type-builds",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=100")
			},
			expectsPayload: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Builds: []model.CreateVersionBuild{
						{Name: "build1", Number: "100"},
					},
				},
			},
			expectsSync:     true,
			expectsDryRun:   false,
			expectsFailFast: true,
		},
		{
			name: "update with explicit sync=false dry-run=true fail-fast=false",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=pkg,version=2.0.0,repo-key=npm-local")
				ctx.AddBoolFlag(commands.SyncFlag, false)
				ctx.AddBoolFlag(commands.DryRunFlag, true)
				ctx.AddBoolFlag(commands.FailFastFlag, false)
			},
			expectsPayload: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "pkg", Version: "2.0.0", Repository: "npm-local"},
					},
				},
			},
			expectsSync:     false,
			expectsDryRun:   true,
			expectsFailFast: false,
		},
		{
			name: "update with spec file",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
			},
			expectsPayload: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "pkg-min", Version: "0.1.0", Repository: "repo-min"},
					},
				},
			},
			expectsSync:     true,
			expectsDryRun:   false,
			expectsFailFast: true,
		},
		{
			name: "update with spec file containing sources and filters",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/filters-spec.json")
			},
			expectsPayload: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "pkg-with-filters", Version: "1.0.0", Repository: "repo-filters"},
					},
				},
				Filters: &model.CreateVersionFilters{
					Included: []*model.CreateVersionSourceFilter{
						{
							PackageType: "docker",
							PackageName: "frontend-*",
						},
					},
					Excluded: []*model.CreateVersionSourceFilter{
						{
							Path: "libs/vulnerable-*.jar",
						},
					},
				},
			},
			expectsSync:     true,
			expectsDryRun:   false,
			expectsFailFast: true,
		},
		{
			name: "update with spec file and spec-vars",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/with-vars-spec.json")
				ctx.AddStringFlag(commands.SpecVarsFlag, "PKG_NAME=my-package;PKG_VERSION=3.0.0;PKG_REPO=npm-prod")
			},
			expectsPayload: &model.UpdateVersionSourcesRequest{
				AddSources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "my-package", Version: "3.0.0", Repository: "npm-prod"},
					},
				},
			},
			expectsSync:     true,
			expectsDryRun:   false,
			expectsFailFast: true,
		},
		{
			name: "spec and source flag together returns error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=pkg,version=1.0.0,repo-key=repo")
			},
			expectsError:  true,
			errorContains: "are not allowed",
		},
		{
			name: "no source flags returns error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsError:  true,
			errorContains: "At least one source flag is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := &components.Context{}
			tt.ctxSetup(ctx)
			ctx.AddStringFlag("url", "https://example.com")
			var actualPayload *model.UpdateVersionSourcesRequest
			var capturedSync, capturedDryRun, capturedFailFast bool
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().UpdateAppVersionSources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, _ string, req *model.UpdateVersionSourcesRequest, sync bool, dryRun bool, failFast bool) error {
						actualPayload = req
						capturedSync = sync
						capturedDryRun = dryRun
						capturedFailFast = failFast
						return nil
					}).Times(1)
			}

			cmd := &updateAppVersionSourcesCommand{
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
				assert.Equal(t, tt.expectsSync, capturedSync, "sync flag mismatch")
				assert.Equal(t, tt.expectsDryRun, capturedDryRun, "dryRun flag mismatch")
				assert.Equal(t, tt.expectsFailFast, capturedFailFast, "failFast flag mismatch")
			}
		})
	}
}
