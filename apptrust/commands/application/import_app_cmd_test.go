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

func TestImportAppCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	content := []byte(`{"applicationKey":"app-key","schemaVersion":1}`)
	filePath := filepath.Join(t.TempDir(), "export.json")
	require.NoError(t, os.WriteFile(filePath, content, 0o600))

	ctx := &components.Context{}
	ctx.Arguments = []string{filePath}
	ctx.AddStringFlag("url", "https://example.com")

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().ImportApplication(gomock.Any(), content).Return(nil).Times(1)

	cmd := &importAppCommand{
		applicationService: mockAppService,
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.NoError(t, err)
}

func TestImportAppCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationEnvelope := []byte(`{"applicationKey":"app-key"}`)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().ImportApplication(gomock.Any(), applicationEnvelope).Return(errors.New("import error")).Times(1)

	cmd := &importAppCommand{
		applicationService:  mockAppService,
		serverDetails:       serverDetails,
		applicationEnvelope: applicationEnvelope,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "import error", err.Error())
}

func TestImportAppCommand_WrongNumberOfArguments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(app, set, nil)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	cmd := &importAppCommand{
		applicationService: mockAppService,
	}

	// Test with no arguments
	context, err := components.ConvertContext(ctx)
	assert.NoError(t, err)

	err = cmd.prepareAndRunCommand(context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Wrong number of arguments")
}

func TestImportAppCommand_MissingFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := &components.Context{}
	ctx.Arguments = []string{filepath.Join(t.TempDir(), "does-not-exist.json")}
	ctx.AddStringFlag("url", "https://example.com")

	cmd := &importAppCommand{
		applicationService: mockapps.NewMockApplicationService(ctrl),
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.Error(t, err)
}
