package application

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	mockapps "github.com/jfrog/jfrog-cli-application/apptrust/service/applications/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

func TestExportAppCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	appKey := "app-key"
	targetPath := filepath.Join(t.TempDir(), "app.json")
	compactJSON := []byte(`{"applicationKey":"app-key","schemaVersion":1}`)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().ExportApplication(gomock.Any(), appKey).Return(compactJSON, nil).Times(1)

	cmd := &exportAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		applicationKey:     appKey,
		targetPath:         targetPath,
	}

	err := cmd.Run()
	assert.NoError(t, err)

	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, compactJSON, got)
}

func TestExportAppCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	appKey := "app-key"

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().ExportApplication(gomock.Any(), appKey).Return(nil, errors.New("export error")).Times(1)

	cmd := &exportAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		applicationKey:     appKey,
		targetPath:         filepath.Join(t.TempDir(), "app.json"),
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "export error", err.Error())
}

func TestExportAppCommand_WrongNumberOfArguments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(app, set, nil)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	cmd := &exportAppCommand{
		applicationService: mockAppService,
	}

	// Test with no arguments
	context, err := components.ConvertContext(ctx)
	assert.NoError(t, err)

	err = cmd.prepareAndRunCommand(context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Wrong number of arguments")
}

func TestResolveExportTargetPath(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "omitted defaults to current directory", target: "", want: "my-app.json"},
		{name: "trailing slash is a directory", target: "exports/", want: filepath.Join("exports", "my-app.json")},
		{name: "no trailing slash is a rename file", target: filepath.Join("a", "b"), want: filepath.Join("a", "b")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveExportTargetPath("my-app", tt.target))
		})
	}
}
