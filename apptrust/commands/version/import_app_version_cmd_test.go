package version

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestImportAppVersionCommand_WrongNumberOfArguments(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
	}{
		{name: "no arguments", arguments: []string{}},
		{name: "single argument", arguments: []string{"app-key"}},
		{name: "too many arguments", arguments: []string{"app-key", "./archive.zip", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				PrintCommandHelp: func(string) error { return nil },
			}
			ctx.Arguments = tt.arguments

			cmd := &importAppVersionCommand{
				versionService: mockversions.NewMockVersionService(ctrl),
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.Error(t, err)
		})
	}
}

func TestImportAppVersionCommand_MissingFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := &components.Context{}
	ctx.Arguments = []string{"app-key", filepath.Join(t.TempDir(), "does-not-exist.zip")}
	ctx.AddStringFlag("url", "https://example.com")

	cmd := &importAppVersionCommand{
		versionService: mockversions.NewMockVersionService(ctrl),
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestImportAppVersionCommand_FlagsSuite(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "archive.zip")
	require.NoError(t, os.WriteFile(archivePath, []byte("not-inspected"), 0o600))

	tests := []struct {
		name           string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsOptions *model.ImportAppVersionOptions
	}{
		{
			name: "defaults to path_mapping",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", archivePath}
			},
			expectsOptions: &model.ImportAppVersionOptions{Mode: model.ImportModePathMapping},
		},
		{
			name: "mapping pattern and target",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", archivePath}
				ctx.AddStringFlag(commands.MappingPatternFlag, "my-repo/(*)")
				ctx.AddStringFlag(commands.MappingTargetFlag, "edge/{1}")
			},
			expectsOptions: &model.ImportAppVersionOptions{
				Mode: model.ImportModePathMapping,
				PathMappings: []model.DistributionPathMapping{
					{Input: "^my-repo/(.*)$", Output: "edge/$1"},
				},
			},
		},
		{
			name: "unpromoted",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", archivePath}
				ctx.AddBoolFlag(commands.UnpromotedFlag, true)
			},
			expectsOptions: &model.ImportAppVersionOptions{Mode: model.ImportModeUnpromoted},
		},
		{
			name: "unpromoted with mapping flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", archivePath}
				ctx.AddBoolFlag(commands.UnpromotedFlag, true)
				ctx.AddStringFlag(commands.MappingPatternFlag, "my-repo/(*)")
				ctx.AddStringFlag(commands.MappingTargetFlag, "edge/{1}")
			},
			expectsError:  true,
			errorContains: "can't be used with",
		},
		{
			name: "only mapping-pattern",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key", archivePath}
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

			var actualOptions *model.ImportAppVersionOptions
			mockVersionService := mockversions.NewMockVersionService(ctrl)
			if !tt.expectsError {
				mockVersionService.EXPECT().ImportAppVersion(gomock.Any(), "app-key", archivePath, gomock.Any()).
					DoAndReturn(func(_ interface{}, _ string, _ string, options *model.ImportAppVersionOptions) ([]byte, error) {
						actualOptions = options
						return []byte(`{"name":"app-key","version":"1.0.0"}`), nil
					}).Times(1)
			}

			cmd := &importAppVersionCommand{
				versionService: mockVersionService,
			}

			err := cmd.prepareAndRunCommand(ctx)
			if tt.expectsError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectsOptions, actualOptions)
		})
	}
}

func TestParseImportAppVersionResponse(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  []byte
		want          importAppVersionResponse
		errorContains string
	}{
		{
			name:         "success",
			responseBody: []byte(`{"name":"app-key","version":"1.0.0"}`),
			want:         importAppVersionResponse{Name: "app-key", Version: "1.0.0"},
		},
		{
			name:          "invalid json",
			responseBody:  []byte(`not-json`),
			errorContains: "failed to parse import response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseImportAppVersionResponse(tt.responseBody)
			if tt.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestImportAppVersionCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	archivePath := "/tmp/archive.zip"
	options := &model.ImportAppVersionOptions{Mode: model.ImportModePathMapping}
	response := []byte(`{"name":"app-key","version":"1.0.0"}`)

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().ImportAppVersion(gomock.Any(), applicationKey, archivePath, options).
		Return(response, nil).Times(1)

	cmd := &importAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		archivePath:    archivePath,
		options:        options,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestImportAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	archivePath := "/tmp/archive.zip"
	options := &model.ImportAppVersionOptions{Mode: model.ImportModePathMapping}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().ImportAppVersion(gomock.Any(), applicationKey, archivePath, options).
		Return(nil, errors.New("import error")).Times(1)

	cmd := &importAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		archivePath:    archivePath,
		options:        options,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "import error", err.Error())
}
