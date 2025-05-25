package version

import (
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &coreconfig.ServerDetails{}
	requestPayload := &model.PromoteAppVersionRequest{
		ApplicationKey: "app",
		Version:        "1.0.0",
		Environment:    "env",
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().PromoteAppVersion(gomock.Any(), requestPayload).
		Return(nil).Times(1)

	cmd := &promoteAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestServerDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &coreconfig.ServerDetails{}
	cmd := &promoteAppVersionCommand{
		serverDetails: serverDetails,
	}

	details, err := cmd.ServerDetails()
	assert.NoError(t, err)
	assert.Equal(t, serverDetails, details)
}

func TestCommandName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &promoteAppVersionCommand{}
	assert.Equal(t, commands.PromoteAppVersion, cmd.CommandName())
}
