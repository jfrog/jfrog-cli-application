package application

import (
	"errors"
	"flag"
	"testing"

	mockapps "github.com/jfrog/jfrog-cli-application/application/service/applications/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"go.uber.org/mock/gomock"
)

func TestDeleteAppCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	appKey := "app-key"

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().DeleteApplication(gomock.Any(), appKey).Return(nil).Times(1)

	cmd := &deleteAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		applicationKey:     appKey,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestDeleteAppCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	appKey := "app-key"

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().DeleteApplication(gomock.Any(), appKey).Return(errors.New("failed to delete application. Status code: 500")).Times(1)

	cmd := &deleteAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		applicationKey:     appKey,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "failed to delete application. Status code: 500", err.Error())
}

func TestDeleteAppCommand_WrongNumberOfArguments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(app, set, nil)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	cmd := &deleteAppCommand{
		applicationService: mockAppService,
	}

	// Test with no arguments
	context, err := components.ConvertContext(ctx)
	assert.NoError(t, err)

	err = cmd.prepareAndRunCommand(context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Wrong number of arguments")
}
