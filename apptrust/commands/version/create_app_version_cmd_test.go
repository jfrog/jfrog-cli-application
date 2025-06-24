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

func TestCreateAppVersionCommand(t *testing.T) {
	tests := []struct {
		name         string
		request      *model.CreateAppVersionRequest
		shouldError  bool
		errorMessage string
	}{
		{
			name: "success",
			request: &model.CreateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Sources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{{
						Type:       "type",
						Name:       "name",
						Version:    "1.0.0",
						Repository: "repo",
					}},
				},
			},
		},
		{
			name:         "context error",
			request:      &model.CreateAppVersionRequest{ApplicationKey: "app-key", Version: "1.0.0", Sources: &model.CreateVersionSources{Packages: []model.CreateVersionPackage{{Type: "type", Name: "name", Version: "1.0.0", Repository: "repo"}}}},
			shouldError:  true,
			errorMessage: "context error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				Arguments: []string{"app-key", "1.0.0"},
			}
			ctx.AddStringFlag("url", "https://example.com")

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if tt.shouldError {
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), tt.request).
					Return(errors.New(tt.errorMessage)).Times(1)
			} else {
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), tt.request).
					Return(nil).Times(1)
			}

			cmd := &createAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  &config.ServerDetails{Url: "https://example.com"},
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

func TestCreateAppVersionCommand_SpecAndFlags_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSpecPath := "./testfiles/test-spec.json"
	ctx := &components.Context{
		Arguments: []string{"app-key", "1.0.0"},
	}
	ctx.AddStringFlag(commands.SpecFlag, testSpecPath)
	ctx.AddStringFlag(commands.PackageNameFlag, "name")
	ctx.AddStringFlag("url", "https://example.com")

	mockVersionService := mockversions.NewMockVersionService(ctrl)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--spec provided")
}

func TestCreateAppVersionCommand_FlagsSuite(t *testing.T) {
	tests := []struct {
		name           string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsPayload *model.CreateAppVersionRequest
	}{
		{
			name: "all flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.TagFlag, "release-tag")
				ctx.AddStringFlag(commands.PackagesFlag, "pkg1:1.2.3,pkg2:2.3.4")
				ctx.AddStringFlag(commands.BuildsFlag, "build1:1.0.0:2024-01-01;build2:2.0.0:2024-02-01")
				ctx.AddStringFlag(commands.ReleaseBundlesFlag, "rb1:1.0.0;rb2:2.0.0")
				ctx.AddStringFlag(commands.SourceVersionFlag, "source-app:3.2.1")
			},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "release-tag",
				Sources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{
						{Name: "pkg1", Version: "1.2.3"},
						{Name: "pkg2", Version: "2.3.4"},
					},
					Builds: []model.CreateVersionBuild{
						{Name: "build1", Number: "1.0.0", Started: "2024-01-01"},
						{Name: "build2", Number: "2.0.0", Started: "2024-02-01"},
					},
					ReleaseBundles: []model.CreateVersionReleaseBundle{
						{Name: "rb1", Version: "1.0.0"},
						{Name: "rb2", Version: "2.0.0"},
					},
					Versions: []model.CreateVersionReference{
						{ApplicationKey: "source-app", Version: "3.2.1"},
					},
				},
			},
		},
		{
			name: "spec only",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "/file1.txt")
			},
			expectsPayload: nil,
			expectsError:   true,
			errorContains:  "no such file or directory",
		},
		{
			name: "spec-vars only",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecVarsFlag, "key1:val1,key2:val2")
			},
			expectsPayload: nil,
			expectsError:   true,
			errorContains:  "At least one source flag is required to create an application version. Please provide one of the following: --spec, --package-name, --builds, --release-bundles, or --source-version.",
		},
		{
			name: "empty flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsPayload: nil,
			expectsError:   true,
			errorContains:  "At least one source flag is required to create an application version. Please provide one of the following: --spec, --package-name, --builds, --release-bundles, or --source-version.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			tt.ctxSetup(ctx)
			ctx.AddStringFlag("url", "https://example.com")

			var actualPayload *model.CreateAppVersionRequest
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.CreateAppVersionRequest) error {
						actualPayload = req
						return nil
					}).Times(1)
			}

			cmd := &createAppVersionCommand{
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

func TestParseBuilds(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic build parsing
	builds, err := cmd.parseBuilds("build1:1.0.0:2024-01-01;build2:2.0.0:2024-02-01")
	assert.NoError(t, err)
	assert.Len(t, builds, 2)
	assert.Equal(t, "build1", builds[0].Name)
	assert.Equal(t, "1.0.0", builds[0].Number)
	assert.Equal(t, "build2", builds[1].Name)
	assert.Equal(t, "2.0.0", builds[1].Number)

	// Test invalid format
	_, err = cmd.parseBuilds("build1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid build format")

	// Test too many parts
	_, err = cmd.parseBuilds("build1:1.0.0:timestamp:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid build format")
}

func TestParseReleaseBundles(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic release bundle parsing
	rbs, err := cmd.parseReleaseBundles("rb1:1.0.0;rb2:2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rbs))
	assert.Equal(t, "rb1", rbs[0].Name)
	assert.Equal(t, "1.0.0", rbs[0].Version)
	assert.Equal(t, "rb2", rbs[1].Name)
	assert.Equal(t, "2.0.0", rbs[1].Version)

	// Test invalid format
	_, err = cmd.parseReleaseBundles("rb1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid release bundle format")

	// Test invalid format with too many parts
	_, err = cmd.parseReleaseBundles("rb1:1.0.0:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid release bundle format")
}

func TestParseSourceVersions(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic source versions parsing
	svs, err := cmd.parseSourceVersions("app1:1.0.0;app2:2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(svs))
	assert.Equal(t, "app1", svs[0].ApplicationKey)
	assert.Equal(t, "1.0.0", svs[0].Version)
	assert.Equal(t, "app2", svs[1].ApplicationKey)
	assert.Equal(t, "2.0.0", svs[1].Version)

	// Test invalid format
	_, err = cmd.parseSourceVersions("app1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source version format")

	// Test invalid format with too many parts
	_, err = cmd.parseSourceVersions("app1:1.0.0:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source version format")
}

func TestCreateAppVersionCommand_SpecFileSuite(t *testing.T) {
	tests := []struct {
		name           string
		specPath       string
		args           []string
		expectsError   bool
		errorContains  string
		expectsPayload *model.CreateAppVersionRequest
	}{
		{
			name:     "minimal spec file",
			specPath: "./testfiles/minimal-spec.json",
			args:     []string{"app-min", "0.1.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-min",
				Version:        "0.1.0",
				Sources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{{
						Type:       "npm",
						Name:       "pkg-min",
						Version:    "0.1.0",
						Repository: "repo-min",
					}},
				},
			},
		},
		{
			name:          "invalid spec file",
			specPath:      "./testfiles/invalid-spec.json",
			args:          []string{"app-invalid", "0.1.0"},
			expectsError:  true,
			errorContains: "invalid character",
		},
		{
			name:     "unknown fields in spec file",
			specPath: "./testfiles/unknown-fields-spec.json",
			args:     []string{"app-unknown", "0.2.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-unknown",
				Version:        "0.2.0",
				Sources: &model.CreateVersionSources{
					Packages: []model.CreateVersionPackage{{
						Type:       "npm",
						Name:       "pkg-unknown",
						Version:    "0.2.0",
						Repository: "repo-unknown",
					}},
				},
			},
		},
		{
			name:          "empty spec file",
			specPath:      "./testfiles/empty-spec.json",
			args:          []string{"app-empty", "0.0.1"},
			expectsError:  true,
			errorContains: "Spec file is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				Arguments: tt.args,
			}
			ctx.AddStringFlag(commands.SpecFlag, tt.specPath)
			ctx.AddStringFlag("url", "https://example.com")

			var actualPayload *model.CreateAppVersionRequest
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.CreateAppVersionRequest) error {
						actualPayload = req
						return nil
					}).Times(1)
			}

			cmd := &createAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  &config.ServerDetails{Url: "https://example.com"},
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
