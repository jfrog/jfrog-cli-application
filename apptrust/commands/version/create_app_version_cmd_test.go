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
	ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=1.0.0")
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
				ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=1.0.0,include_deps=true;name=build2,id=2.0.0,include_deps=false")
				ctx.AddStringFlag(commands.SourceTypeReleaseBundlesFlag, "name=rb1,version=1.0.0;name=rb2,version=2.0.0")
				ctx.AddStringFlag(commands.SourceTypeApplicationVersionsFlag, "application-key=source-app,version=3.2.1")
			},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "release-tag",
				Sources: &model.CreateVersionSources{
					Builds: []model.CreateVersionBuild{
						{Name: "build1", Number: "1.0.0", IncludeDependencies: true},
						{Name: "build2", Number: "2.0.0", IncludeDependencies: false},
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
			errorContains:  "At least one source flag is required to create an application version. Please provide --spec or at least one of the following: --source-type-builds, --source-type-release-bundles, --source-type-application-versions.",
		},
		{
			name: "empty flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsPayload: nil,
			expectsError:   true,
			errorContains:  "At least one source flag is required to create an application version. Please provide --spec or at least one of the following: --source-type-builds, --source-type-release-bundles, --source-type-application-versions.",
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

	tests := []struct {
		name           string
		input          string
		expectError    bool
		errorContains  string
		expectedBuilds []model.CreateVersionBuild
	}{
		{
			name:        "multiple builds",
			input:       "name=build1,id=1.0.0,include_deps=true;name=build2,id=2.0.0,include_deps=false;name=build3,id=3.0.0",
			expectError: false,
			expectedBuilds: []model.CreateVersionBuild{
				{Name: "build1", Number: "1.0.0", IncludeDependencies: true},
				{Name: "build2", Number: "2.0.0", IncludeDependencies: false},
				{Name: "build3", Number: "3.0.0", IncludeDependencies: false},
			},
		},
		{
			name:           "empty string",
			input:          "",
			expectError:    false,
			expectedBuilds: nil,
		},
		{
			name:          "missing name field",
			input:         "id=1.0.0",
			expectError:   true,
			errorContains: "missing required field: name",
		},
		{
			name:          "missing id field",
			input:         "name=build1",
			expectError:   true,
			errorContains: "missing required field: id",
		},
		{
			name:          "invalid format",
			input:         "build1",
			expectError:   true,
			errorContains: "invalid build format",
		},
		{
			name:          "invalid include_deps value",
			input:         "name=build1,id=1.0.0,include_deps=invalid",
			expectError:   true,
			errorContains: "invalid build format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builds, err := cmd.parseBuilds(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBuilds, builds)
			}
		})
	}
}

func TestParseReleaseBundles(t *testing.T) {
	cmd := &createAppVersionCommand{}

	tests := []struct {
		name                   string
		input                  string
		expectError            bool
		errorContains          string
		expectedReleaseBundles []model.CreateVersionReleaseBundle
	}{
		{
			name:        "multiple release bundles",
			input:       "name=rb1,version=1.0.0;name=rb2,version=2.0.0",
			expectError: false,
			expectedReleaseBundles: []model.CreateVersionReleaseBundle{
				{Name: "rb1", Version: "1.0.0"},
				{Name: "rb2", Version: "2.0.0"},
			},
		},
		{
			name:                   "empty string",
			input:                  "",
			expectError:            false,
			expectedReleaseBundles: nil,
		},
		{
			name:          "missing name field",
			input:         "version=1.0.0",
			expectError:   true,
			errorContains: "missing required field: name",
		},
		{
			name:          "missing version field",
			input:         "name=rb1",
			expectError:   true,
			errorContains: "missing required field: version",
		},
		{
			name:          "invalid format",
			input:         "rb1",
			expectError:   true,
			errorContains: "invalid release bundle format",
		},
		{
			name:        "with project-key and repo-key",
			input:       "name=rb1,version=1.0.0,project-key=proj1,repo-key=repo1",
			expectError: false,
			expectedReleaseBundles: []model.CreateVersionReleaseBundle{
				{Name: "rb1", Version: "1.0.0", ProjectKey: "proj1", RepositoryKey: "repo1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rbs, err := cmd.parseReleaseBundles(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedReleaseBundles, rbs)
			}
		})
	}
}

func TestParseSourceVersions(t *testing.T) {
	cmd := &createAppVersionCommand{}

	tests := []struct {
		name                   string
		input                  string
		expectError            bool
		errorContains          string
		expectedSourceVersions []model.CreateVersionReference
	}{
		{
			name:        "multiple source versions",
			input:       "application-key=app1,version=1.0.0;application-key=app2,version=2.0.0",
			expectError: false,
			expectedSourceVersions: []model.CreateVersionReference{
				{ApplicationKey: "app1", Version: "1.0.0"},
				{ApplicationKey: "app2", Version: "2.0.0"},
			},
		},
		{
			name:                   "empty string",
			input:                  "",
			expectError:            false,
			expectedSourceVersions: nil,
		},
		{
			name:          "missing application-key field",
			input:         "version=1.0.0",
			expectError:   true,
			errorContains: "missing required field: application-key",
		},
		{
			name:          "missing version field",
			input:         "application-key=app1",
			expectError:   true,
			errorContains: "missing required field: version",
		},
		{
			name:          "invalid format",
			input:         "app1",
			expectError:   true,
			errorContains: "invalid application version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svs, err := cmd.parseSourceVersions(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSourceVersions, svs)
			}
		})
	}
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
		{
			name:     "artifacts spec file",
			specPath: "./testfiles/artifacts-spec.json",
			args:     []string{"app-artifacts", "1.0.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-artifacts",
				Version:        "1.0.0",
				Sources: &model.CreateVersionSources{
					Artifacts: []model.CreateVersionArtifact{
						{
							Path:   "repo/path/to/artifact1.jar",
							SHA256: "abc123def456",
						},
						{
							Path: "repo/path/to/artifact2.war",
						},
					},
				},
			},
		},
		{
			name:     "artifacts only spec file",
			specPath: "./testfiles/artifacts-only-spec.json",
			args:     []string{"app-single-artifact", "2.0.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-single-artifact",
				Version:        "2.0.0",
				Sources: &model.CreateVersionSources{
					Artifacts: []model.CreateVersionArtifact{
						{
							Path:   "repo/path/to/single-artifact.jar",
							SHA256: "1234567890abcdef",
						},
					},
				},
			},
		},
		{
			name:     "mixed sources with artifacts spec file",
			specPath: "./testfiles/mixed-sources-spec.json",
			args:     []string{"app-mixed", "3.0.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-mixed",
				Version:        "3.0.0",
				Sources: &model.CreateVersionSources{
					Artifacts: []model.CreateVersionArtifact{
						{
							Path:   "repo/path/to/artifact.jar",
							SHA256: "abc123",
						},
					},
					Packages: []model.CreateVersionPackage{
						{
							Type:       "npm",
							Name:       "my-package",
							Version:    "1.0.0",
							Repository: "npm-repo",
						},
					},
					Builds: []model.CreateVersionBuild{
						{
							Name:   "build1",
							Number: "42",
						},
					},
				},
			},
		},
		{
			name:     "all sources spec file",
			specPath: "./testfiles/all-sources-spec.json",
			args:     []string{"app-all-sources", "5.0.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-all-sources",
				Version:        "5.0.0",
				Sources: &model.CreateVersionSources{
					Artifacts: []model.CreateVersionArtifact{
						{
							Path:   "repo/path/to/app.jar",
							SHA256: "abc123def456789",
						},
						{
							Path: "repo/path/to/lib.war",
						},
					},
					Packages: []model.CreateVersionPackage{
						{
							Type:       "npm",
							Name:       "my-package",
							Version:    "1.2.3",
							Repository: "npm-local",
						},
						{
							Type:       "docker",
							Name:       "my-docker-image",
							Version:    "2.0.0",
							Repository: "docker-local",
						},
					},
					Builds: []model.CreateVersionBuild{
						{
							Name:                "my-build",
							Number:              "123",
							IncludeDependencies: true,
						},
						{
							Name:                "another-build",
							Number:              "456",
							RepositoryKey:       "build-info",
							IncludeDependencies: false,
						},
					},
					ReleaseBundles: []model.CreateVersionReleaseBundle{
						{
							Name:          "my-release-bundle",
							Version:       "1.0.0",
							ProjectKey:    "my-project",
							RepositoryKey: "rb-repo",
						},
						{
							Name:    "another-bundle",
							Version: "2.0.0",
						},
					},
					Versions: []model.CreateVersionReference{
						{
							ApplicationKey: "dependency-app-1",
							Version:        "3.0.0",
						},
						{
							ApplicationKey: "dependency-app-2",
							Version:        "4.5.6",
						},
					},
				},
			},
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

func TestValidateCreateAppVersionContext(t *testing.T) {
	tests := []struct {
		name          string
		ctxSetup      func(*components.Context)
		expectError   bool
		errorContains string
	}{
		{
			name: "no source flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectError:   true,
			errorContains: "At least one source flag is required",
		},
		{
			name: "valid context with builds flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=1.0.0")
			},
			expectError: false,
		},
		{
			name: "valid context with spec flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			tt.ctxSetup(ctx)

			err := validateCreateAppVersionContext(ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNoSpecAndFlagsTogether(t *testing.T) {
	tests := []struct {
		name          string
		ctxSetup      func(*components.Context)
		expectError   bool
		errorContains string
	}{
		{
			name: "spec flag with builds flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=1.0.0")
			},
			expectError:   true,
			errorContains: "--spec provided",
		},
		{
			name: "spec flag with release bundles flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.SourceTypeReleaseBundlesFlag, "name=rb1,version=1.0.0")
			},
			expectError:   true,
			errorContains: "--spec provided",
		},
		{
			name: "spec flag with application versions flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.SourceTypeApplicationVersionsFlag, "application-key=app1,version=1.0.0")
			},
			expectError:   true,
			errorContains: "--spec provided",
		},
		{
			name: "spec flag only",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
			},
			expectError: false,
		},
		{
			name: "other flags only",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=1.0.0")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			tt.ctxSetup(ctx)

			err := validateNoSpecAndFlagsTogether(ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequiredFieldsInMap(t *testing.T) {
	tests := []struct {
		name           string
		inputMap       map[string]string
		requiredFields []string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "nil map",
			inputMap:       nil,
			requiredFields: []string{"field1", "field2"},
			expectError:    true,
			errorContains:  "missing required fields: field1, field2",
		},
		{
			name:           "missing field",
			inputMap:       map[string]string{"field1": "value1"},
			requiredFields: []string{"field1", "field2"},
			expectError:    true,
			errorContains:  "missing required field: field2",
		},
		{
			name:           "all required fields present",
			inputMap:       map[string]string{"field1": "value1", "field2": "value2", "extra": "value3"},
			requiredFields: []string{"field1", "field2"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredFieldsInMap(tt.inputMap, tt.requiredFields...)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
