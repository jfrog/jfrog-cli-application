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
				Draft:          false,
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
			request:      &model.CreateAppVersionRequest{ApplicationKey: "app-key", Version: "1.0.0", Draft: false, Sources: &model.CreateVersionSources{Packages: []model.CreateVersionPackage{{Type: "type", Name: "name", Version: "1.0.0", Repository: "repo"}}}},
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
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), tt.request, true).
					Return(errors.New(tt.errorMessage)).Times(1)
			} else {
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), tt.request, true).
					Return(nil).Times(1)
			}

			cmd := &createAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  &config.ServerDetails{Url: "https://example.com"},
				requestPayload: tt.request,
				sync:           true,
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
				ctx.AddBoolFlag(commands.DraftFlag, true)
				ctx.AddStringFlag(commands.SourceTypeBuildsFlag, "name=build1,id=1.0.0,include-deps=true,repo-key=build-info-repo,started=2024-01-15T10:30:00Z;name=build2,id=2.0.0,include-deps=false")
				ctx.AddStringFlag(commands.SourceTypeReleaseBundlesFlag, "name=rb1,version=1.0.0;name=rb2,version=2.0.0")
				ctx.AddStringFlag(commands.SourceTypeApplicationVersionsFlag, "application-key=source-app,version=3.2.1")
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=pkg1,version=1.0.0,repo-key=repo1;type=docker,name=pkg2,version=2.0.0,repo-key=repo2")
				ctx.AddStringFlag(commands.SourceTypeArtifactsFlag, "path=repo/path/to/artifact1.jar,sha256=abc123;path=repo/path/to/artifact2.war")
			},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-key",
				Version:        "1.0.0",
				Tag:            "release-tag",
				Draft:          true,
				Sources: &model.CreateVersionSources{
					Builds: []model.CreateVersionBuild{
						{Name: "build1", Number: "1.0.0", IncludeDependencies: true, RepositoryKey: "build-info-repo", Started: "2024-01-15T10:30:00Z"},
						{Name: "build2", Number: "2.0.0", IncludeDependencies: false},
					},
					ReleaseBundles: []model.CreateVersionReleaseBundle{
						{Name: "rb1", Version: "1.0.0"},
						{Name: "rb2", Version: "2.0.0"},
					},
					Versions: []model.CreateVersionReference{
						{ApplicationKey: "source-app", Version: "3.2.1"},
					},
					Packages: []model.CreateVersionPackage{
						{Type: "npm", Name: "pkg1", Version: "1.0.0", Repository: "repo1"},
						{Type: "docker", Name: "pkg2", Version: "2.0.0", Repository: "repo2"},
					},
					Artifacts: []model.CreateVersionArtifact{
						{Path: "repo/path/to/artifact1.jar", SHA256: "abc123"},
						{Path: "repo/path/to/artifact2.war"},
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
			errorContains:  "At least one source flag is required to create an application version. Please provide --spec or at least one of the following: --source-type-builds, --source-type-release-bundles, --source-type-application-versions, --source-type-packages, --source-type-artifacts.",
		},
		{
			name: "empty flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
			},
			expectsPayload: nil,
			expectsError:   true,
			errorContains:  "At least one source flag is required to create an application version. Please provide --spec or at least one of the following: --source-type-builds, --source-type-release-bundles, --source-type-application-versions, --source-type-packages, --source-type-artifacts.",
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
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.CreateAppVersionRequest, _ bool) error {
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
	tests := []struct {
		name           string
		input          string
		expectError    bool
		errorContains  string
		expectedBuilds []model.CreateVersionBuild
	}{
		{
			name:        "multiple builds",
			input:       "name=build1,id=1.0.0,include-deps=true;name=build2,id=2.0.0,started=2024-02-20T14:45:00Z,include-deps=false;name=build3,id=3.0.0,repo-key=custom-build-repo",
			expectError: false,
			expectedBuilds: []model.CreateVersionBuild{
				{Name: "build1", Number: "1.0.0", IncludeDependencies: true},
				{Name: "build2", Number: "2.0.0", Started: "2024-02-20T14:45:00Z", IncludeDependencies: false},
				{Name: "build3", Number: "3.0.0", IncludeDependencies: false, RepositoryKey: "custom-build-repo"},
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
			name:          "invalid include-deps value",
			input:         "name=build1,id=1.0.0,include-deps=invalid",
			expectError:   true,
			errorContains: "invalid build format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builds, err := parseBuilds(tt.input)
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
			rbs, err := parseReleaseBundles(tt.input)
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
			svs, err := parseSourceVersions(tt.input)
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

func TestParsePackages(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectError      bool
		errorContains    string
		expectedPackages []model.CreateVersionPackage
	}{
		{
			name:        "multiple packages",
			input:       "type=npm,name=pkg1,version=1.0.0,repo-key=repo1;type=docker,name=pkg2,version=2.0.0,repo-key=repo2",
			expectError: false,
			expectedPackages: []model.CreateVersionPackage{
				{Type: "npm", Name: "pkg1", Version: "1.0.0", Repository: "repo1"},
				{Type: "docker", Name: "pkg2", Version: "2.0.0", Repository: "repo2"},
			},
		},
		{
			name:             "empty string",
			input:            "",
			expectError:      false,
			expectedPackages: nil,
		},
		{
			name:          "missing type field",
			input:         "name=pkg1,version=1.0.0,repo-key=repo1",
			expectError:   true,
			errorContains: "missing required field: type",
		},
		{
			name:          "missing name field",
			input:         "type=npm,version=1.0.0,repo-key=repo1",
			expectError:   true,
			errorContains: "missing required field: name",
		},
		{
			name:          "missing version field",
			input:         "type=npm,name=pkg1,repo-key=repo1",
			expectError:   true,
			errorContains: "missing required field: version",
		},
		{
			name:          "missing repo-key field",
			input:         "type=npm,name=pkg1,version=1.0.0",
			expectError:   true,
			errorContains: "missing required field: repo-key",
		},
		{
			name:          "invalid format",
			input:         "pkg1",
			expectError:   true,
			errorContains: "invalid package format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages, err := parsePackages(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPackages, packages)
			}
		})
	}
}

func TestParseArtifacts(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectError       bool
		errorContains     string
		expectedArtifacts []model.CreateVersionArtifact
	}{
		{
			name:        "multiple artifacts",
			input:       "path=repo/path/to/artifact1.jar,sha256=abc123def456;path=repo/path/to/artifact2.war",
			expectError: false,
			expectedArtifacts: []model.CreateVersionArtifact{
				{Path: "repo/path/to/artifact1.jar", SHA256: "abc123def456"},
				{Path: "repo/path/to/artifact2.war"},
			},
		},
		{
			name:              "empty string",
			input:             "",
			expectError:       false,
			expectedArtifacts: nil,
		},
		{
			name:          "missing path field",
			input:         "sha256=abc123def456",
			expectError:   true,
			errorContains: "missing required field: path",
		},
		{
			name:          "invalid format",
			input:         "artifact1.jar",
			expectError:   true,
			errorContains: "invalid artifact format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifacts, err := parseArtifacts(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedArtifacts, artifacts)
			}
		})
	}
}

func TestCreateAppVersionCommand_SpecFileSuite(t *testing.T) {
	tests := []struct {
		name           string
		specPath       string
		args           []string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsPayload *model.CreateAppVersionRequest
		expectsSync    *bool
	}{
		{
			name:     "minimal spec file",
			specPath: "./testfiles/minimal-spec.json",
			args:     []string{"app-min", "0.1.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-min",
				Version:        "0.1.0",
				Draft:          false,
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
				Draft:          false,
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
				Draft:          false,
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
			name:     "all sources spec file",
			specPath: "./testfiles/all-sources-spec.json",
			args:     []string{"app-all-sources", "5.0.0"},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-all-sources",
				Version:        "5.0.0",
				Draft:          false,
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
		{
			name:     "spec file with draft flag",
			specPath: "./testfiles/minimal-spec.json",
			args:     []string{"app-draft-spec", "2.0.0"},
			ctxSetup: func(ctx *components.Context) {
				ctx.AddBoolFlag(commands.DraftFlag, true)
			},
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-draft-spec",
				Version:        "2.0.0",
				Draft:          true,
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
			name:     "spec file with all flags",
			specPath: "./testfiles/all-sources-spec.json",
			args:     []string{"app-all-flags", "7.0.0"},
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.TagFlag, "v7-release")
				ctx.AddBoolFlag(commands.DraftFlag, true)
				ctx.AddBoolFlag(commands.SyncFlag, false)
			},
			expectsSync: boolPtr(false),
			expectsPayload: &model.CreateAppVersionRequest{
				ApplicationKey: "app-all-flags",
				Version:        "7.0.0",
				Tag:            "v7-release",
				Draft:          true,
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
			if tt.ctxSetup != nil {
				tt.ctxSetup(ctx)
			}

			var actualPayload *model.CreateAppVersionRequest
			var capturedSync bool
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.CreateAppVersionRequest, sync bool) error {
						actualPayload = req
						capturedSync = sync
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
				if tt.expectsSync != nil {
					assert.Equal(t, *tt.expectsSync, capturedSync)
				}
			}
		})
	}
}

func TestCreateAppVersionCommand_SyncFlagSuite(t *testing.T) {
	tests := []struct {
		name        string
		setSync     *bool
		expectsSync bool
	}{
		{
			name:        "defaults to true when not set",
			setSync:     nil,
			expectsSync: true,
		},
		{
			name:        "explicit true",
			setSync:     boolPtr(true),
			expectsSync: true,
		},
		{
			name:        "explicit false",
			setSync:     boolPtr(false),
			expectsSync: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				Arguments: []string{"app-sync-test", "1.0.0"},
			}
			ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
			ctx.AddStringFlag("url", "https://example.com")
			if tt.setSync != nil {
				ctx.AddBoolFlag(commands.SyncFlag, *tt.setSync)
			}

			var capturedSync bool
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ interface{}, req *model.CreateAppVersionRequest, sync bool) error {
					capturedSync = sync
					return nil
				}).Times(1)

			cmd := &createAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  &config.ServerDetails{Url: "https://example.com"},
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectsSync, capturedSync)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
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
			name: "valid context with packages flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=pkg1,version=1.0.0,repo-key=repo1")
			},
			expectError: false,
		},
		{
			name: "valid context with artifacts flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypeArtifactsFlag, "path=repo/path/to/artifact1.jar")
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
			name: "spec flag with packages flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=pkg1,version=1.0.0,repo-key=repo1")
			},
			expectError:   true,
			errorContains: "--spec provided",
		},
		{
			name: "spec flag with artifacts flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.SourceTypeArtifactsFlag, "path=repo/path/to/artifact1.jar")
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

func TestValidateNoSpecAndFlagsTogether_WithFilterFlags(t *testing.T) {
	tests := []struct {
		name          string
		ctxSetup      func(*components.Context)
		expectError   bool
		errorContains string
	}{
		{
			name: "spec with include filter flag - should error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker, name=frontend-*")
			},
			expectError:   true,
			errorContains: "filter flags",
		},
		{
			name: "spec with exclude filter flag - should error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.ExcludeFilterFlag, "filter_type=package, name=*-dev")
			},
			expectError:   true,
			errorContains: "filter flags",
		},
		{
			name: "spec with both filter flags - should error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker")
				ctx.AddStringFlag(commands.ExcludeFilterFlag, "filter_type=package, name=*-dev")
			},
			expectError:   true,
			errorContains: "filter flags",
		},
		{
			name: "spec without filter flags - should not error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
			},
			expectError: false,
		},
		{
			name: "no spec with filter flags - should not error",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=test,version=1.0.0,repo-key=repo")
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker")
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

func TestParseFilters(t *testing.T) {
	tests := []struct {
		name            string
		input           []string
		expectError     bool
		errorContains   string
		expectedFilters []*model.CreateVersionSourceFilter
	}{
		{
			name:  "package filter with type and name",
			input: []string{"filter_type=package, type=docker, name=frontend-*"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{PackageType: "docker", PackageName: "frontend-*"},
			},
		},
		{
			name:  "package filter with all fields",
			input: []string{"filter_type=package, type=npm, name=my-package, version=1.0.0"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{PackageType: "npm", PackageName: "my-package", PackageVersion: "1.0.0"},
			},
		},
		{
			name:  "package filter with only name",
			input: []string{"filter_type=package, name=*-dev"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{PackageName: "*-dev"},
			},
		},
		{
			name:  "package filter with only version",
			input: []string{"filter_type=package, version=3.*"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{PackageVersion: "3.*"},
			},
		},
		{
			name:  "artifact filter with path",
			input: []string{"filter_type=artifact, path=libs/*.jar"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{Path: "libs/*.jar"},
			},
		},
		{
			name:  "artifact filter with path and sha256",
			input: []string{"filter_type=artifact, path=libs/artifact.jar, sha256=a1b2c3d4e5f6"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{Path: "libs/artifact.jar", SHA256: "a1b2c3d4e5f6"},
			},
		},
		{
			name:  "artifact filter with only sha256",
			input: []string{"filter_type=artifact, sha256=a1b2c3d4e5f6789012345678901234567890123456789012345678901267890"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{SHA256: "a1b2c3d4e5f6789012345678901234567890123456789012345678901267890"},
			},
		},
		{
			name:  "multiple filters - package and artifact",
			input: []string{"filter_type=package, type=docker, name=frontend-*", "filter_type=artifact, path=libs/*.jar"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{PackageType: "docker", PackageName: "frontend-*"},
				{Path: "libs/*.jar"},
			},
		},
		{
			name:          "missing filter_type",
			input:         []string{"type=docker, name=frontend-*"},
			expectError:   true,
			errorContains: "missing 'filter_type' field",
		},
		{
			name:          "invalid filter_type",
			input:         []string{"filter_type=invalid, type=docker"},
			expectError:   true,
			errorContains: "invalid filter_type 'invalid'",
		},
		{
			name:          "package filter with no fields",
			input:         []string{"filter_type=package"},
			expectError:   true,
			errorContains: "at least one of 'type', 'name', or 'version' must be specified",
		},
		{
			name:          "artifact filter with no fields",
			input:         []string{"filter_type=artifact"},
			expectError:   true,
			errorContains: "at least one of 'path' or 'sha256' must be specified",
		},
		{
			name:          "invalid format - missing equals",
			input:         []string{"filter_type=package type=docker"},
			expectError:   true,
			errorContains: "invalid filter_type",
		},
		{
			name:            "empty input",
			input:           []string{},
			expectedFilters: []*model.CreateVersionSourceFilter{},
		},
		{
			name:  "multiple package filters",
			input: []string{"filter_type=package, type=docker, name=frontend-*", "filter_type=package, version=3.*"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{PackageType: "docker", PackageName: "frontend-*"},
				{PackageVersion: "3.*"},
			},
		},
		{
			name:  "multiple artifact filters",
			input: []string{"filter_type=artifact, path=libs/*.jar", "filter_type=artifact, path=libs/vulnerable-lib-1.2.3.jar"},
			expectedFilters: []*model.CreateVersionSourceFilter{
				{Path: "libs/*.jar"},
				{Path: "libs/vulnerable-lib-1.2.3.jar"},
			},
		},
		{
			name:          "error in second filter",
			input:         []string{"filter_type=package, type=docker, name=frontend-*", "filter_type=package"},
			expectError:   true,
			errorContains: "at least one of 'type', 'name', or 'version' must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := parseFilters(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedFilters), len(filters))
				for i, expected := range tt.expectedFilters {
					assert.Equal(t, expected.PackageType, filters[i].PackageType)
					assert.Equal(t, expected.PackageName, filters[i].PackageName)
					assert.Equal(t, expected.PackageVersion, filters[i].PackageVersion)
					assert.Equal(t, expected.Path, filters[i].Path)
					assert.Equal(t, expected.SHA256, filters[i].SHA256)
				}
			}
		})
	}
}

func TestBuildFiltersFromFlags(t *testing.T) {
	tests := []struct {
		name            string
		ctxSetup        func(*components.Context)
		expectError     bool
		errorContains   string
		expectedFilters *model.CreateVersionFilters
	}{
		{
			name: "no filters",
			ctxSetup: func(ctx *components.Context) {
			},
			expectedFilters: nil,
		},
		{
			name: "single include filter",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker, name=frontend-*")
			},
			expectedFilters: &model.CreateVersionFilters{
				Included: []*model.CreateVersionSourceFilter{
					{PackageType: "docker", PackageName: "frontend-*"},
				},
			},
		},
		{
			name: "single exclude filter",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.ExcludeFilterFlag, "filter_type=package, name=*-dev")
			},
			expectedFilters: &model.CreateVersionFilters{
				Excluded: []*model.CreateVersionSourceFilter{
					{PackageName: "*-dev"},
				},
			},
		},
		{
			name: "include and exclude filters",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker, name=frontend-*")
				ctx.AddStringFlag(commands.ExcludeFilterFlag, "filter_type=package, name=*-dev")
			},
			expectedFilters: &model.CreateVersionFilters{
				Included: []*model.CreateVersionSourceFilter{
					{PackageType: "docker", PackageName: "frontend-*"},
				},
				Excluded: []*model.CreateVersionSourceFilter{
					{PackageName: "*-dev"},
				},
			},
		},
		{
			name: "multiple include and exclude filters",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker, name=frontend-*; filter_type=artifact, path=libs/*.jar")
				ctx.AddStringFlag(commands.ExcludeFilterFlag, "filter_type=package, name=*-dev; filter_type=package, name=*versions*")
			},
			expectedFilters: &model.CreateVersionFilters{
				Included: []*model.CreateVersionSourceFilter{
					{PackageType: "docker", PackageName: "frontend-*"},
					{Path: "libs/*.jar"},
				},
				Excluded: []*model.CreateVersionSourceFilter{
					{PackageName: "*-dev"},
					{PackageName: "*versions*"},
				},
			},
		},
		{
			name: "invalid include filter",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package")
			},
			expectError:   true,
			errorContains: "at least one of 'type', 'name', or 'version' must be specified",
		},
		{
			name: "invalid exclude filter",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.ExcludeFilterFlag, "filter_type=artifact")
			},
			expectError:   true,
			errorContains: "at least one of 'path' or 'sha256' must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			tt.ctxSetup(ctx)

			filters, err := buildFiltersFromFlags(ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectedFilters == nil {
					assert.Nil(t, filters)
				} else {
					assert.NotNil(t, filters)
					if tt.expectedFilters.Included != nil {
						assert.Equal(t, len(tt.expectedFilters.Included), len(filters.Included))
						for i, expected := range tt.expectedFilters.Included {
							assert.Equal(t, expected.PackageType, filters.Included[i].PackageType)
							assert.Equal(t, expected.PackageName, filters.Included[i].PackageName)
							assert.Equal(t, expected.PackageVersion, filters.Included[i].PackageVersion)
							assert.Equal(t, expected.Path, filters.Included[i].Path)
							assert.Equal(t, expected.SHA256, filters.Included[i].SHA256)
						}
					}
					if tt.expectedFilters.Excluded != nil {
						assert.Equal(t, len(tt.expectedFilters.Excluded), len(filters.Excluded))
						for i, expected := range tt.expectedFilters.Excluded {
							assert.Equal(t, expected.PackageType, filters.Excluded[i].PackageType)
							assert.Equal(t, expected.PackageName, filters.Excluded[i].PackageName)
							assert.Equal(t, expected.PackageVersion, filters.Excluded[i].PackageVersion)
							assert.Equal(t, expected.Path, filters.Excluded[i].Path)
							assert.Equal(t, expected.SHA256, filters.Excluded[i].SHA256)
						}
					}
				}
			}
		})
	}
}

func TestLoadFromSpec_WithFilters(t *testing.T) {
	tests := []struct {
		name            string
		specPath        string
		expectError     bool
		errorContains   string
		expectedSources *model.CreateVersionSources
		expectedFilters *model.CreateVersionFilters
	}{
		{
			name:     "spec with included filters",
			specPath: "./testfiles/filters-spec.json",
			expectedSources: &model.CreateVersionSources{
				Packages: []model.CreateVersionPackage{
					{
						Type:       "npm",
						Name:       "pkg-with-filters",
						Version:    "1.0.0",
						Repository: "repo-filters",
					},
				},
			},
			expectedFilters: &model.CreateVersionFilters{
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
		{
			name:     "spec without filters",
			specPath: "./testfiles/minimal-spec.json",
			expectedSources: &model.CreateVersionSources{
				Packages: []model.CreateVersionPackage{
					{
						Type:       "npm",
						Name:       "pkg-min",
						Version:    "0.1.0",
						Repository: "repo-min",
					},
				},
			},
			expectedFilters: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			ctx.AddStringFlag(commands.SpecFlag, tt.specPath)
			ctx.AddStringFlag("url", "https://example.com")

			sources, filters, err := loadSourcesFromSpec(ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSources, sources)
				if tt.expectedFilters == nil {
					assert.Nil(t, filters)
				} else {
					assert.NotNil(t, filters)
					if tt.expectedFilters.Included != nil {
						assert.Equal(t, len(tt.expectedFilters.Included), len(filters.Included))
						for i, expected := range tt.expectedFilters.Included {
							assert.Equal(t, expected.PackageType, filters.Included[i].PackageType)
							assert.Equal(t, expected.PackageName, filters.Included[i].PackageName)
							assert.Equal(t, expected.PackageVersion, filters.Included[i].PackageVersion)
							assert.Equal(t, expected.Path, filters.Included[i].Path)
							assert.Equal(t, expected.SHA256, filters.Included[i].SHA256)
						}
					}
					if tt.expectedFilters.Excluded != nil {
						assert.Equal(t, len(tt.expectedFilters.Excluded), len(filters.Excluded))
						for i, expected := range tt.expectedFilters.Excluded {
							assert.Equal(t, expected.PackageType, filters.Excluded[i].PackageType)
							assert.Equal(t, expected.PackageName, filters.Excluded[i].PackageName)
							assert.Equal(t, expected.PackageVersion, filters.Excluded[i].PackageVersion)
							assert.Equal(t, expected.Path, filters.Excluded[i].Path)
							assert.Equal(t, expected.SHA256, filters.Excluded[i].SHA256)
						}
					}
				}
			}
		})
	}
}

func TestBuildRequestPayload_Filters(t *testing.T) {
	cmd := &createAppVersionCommand{}

	tests := []struct {
		name            string
		ctxSetup        func(*components.Context)
		expectError     bool
		errorContains   string
		expectedFilters *model.CreateVersionFilters
	}{
		{
			name: "filters from spec when spec flag is set",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/filters-spec.json")
				ctx.AddStringFlag("url", "https://example.com")
			},
			expectedFilters: &model.CreateVersionFilters{
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
		{
			name: "filters from flags when spec flag is not set",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SourceTypePackagesFlag, "type=npm,name=test,version=1.0.0,repo-key=repo")
				ctx.AddStringFlag(commands.IncludeFilterFlag, "filter_type=package, type=docker, name=frontend-*")
				ctx.AddStringFlag("url", "https://example.com")
			},
			expectedFilters: &model.CreateVersionFilters{
				Included: []*model.CreateVersionSourceFilter{
					{
						PackageType: "docker",
						PackageName: "frontend-*",
					},
				},
			},
		},
		{
			name: "no filters when spec has no filters and no filter flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", "1.0.0"}
				ctx.AddStringFlag(commands.SpecFlag, "./testfiles/minimal-spec.json")
				ctx.AddStringFlag("url", "https://example.com")
			},
			expectedFilters: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			tt.ctxSetup(ctx)

			payload, err := cmd.buildRequestPayload(ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payload)
				if tt.expectedFilters == nil {
					assert.Nil(t, payload.Filters)
				} else {
					assert.NotNil(t, payload.Filters)
					if tt.expectedFilters.Included != nil {
						assert.Equal(t, len(tt.expectedFilters.Included), len(payload.Filters.Included))
						for i, expected := range tt.expectedFilters.Included {
							assert.Equal(t, expected.PackageType, payload.Filters.Included[i].PackageType)
							assert.Equal(t, expected.PackageName, payload.Filters.Included[i].PackageName)
							assert.Equal(t, expected.PackageVersion, payload.Filters.Included[i].PackageVersion)
							assert.Equal(t, expected.Path, payload.Filters.Included[i].Path)
							assert.Equal(t, expected.SHA256, payload.Filters.Included[i].SHA256)
						}
					}
					if tt.expectedFilters.Excluded != nil {
						assert.Equal(t, len(tt.expectedFilters.Excluded), len(payload.Filters.Excluded))
						for i, expected := range tt.expectedFilters.Excluded {
							assert.Equal(t, expected.PackageType, payload.Filters.Excluded[i].PackageType)
							assert.Equal(t, expected.PackageName, payload.Filters.Excluded[i].PackageName)
							assert.Equal(t, expected.PackageVersion, payload.Filters.Excluded[i].PackageVersion)
							assert.Equal(t, expected.Path, payload.Filters.Excluded[i].Path)
							assert.Equal(t, expected.SHA256, payload.Filters.Excluded[i].SHA256)
						}
					}
				}
			}
		})
	}
}
