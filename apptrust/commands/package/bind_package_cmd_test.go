package packagecmds

import (
	"errors"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	mockpackages "github.com/jfrog/jfrog-cli-application/apptrust/service/packages/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBindPackageCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.BindPackageRequest{
		ApplicationKey: "app-key",
		Type:           "npm",
		Name:           "test-package",
		Versions:       []string{"1.0.0"},
	}

	mockPackageService := mockpackages.NewMockPackageService(ctrl)
	mockPackageService.EXPECT().BindPackage(gomock.Any(), requestPayload).
		Return(nil).Times(1)

	cmd := &bindPackageCommand{
		packageService: mockPackageService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestBindPackageCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.BindPackageRequest{
		ApplicationKey: "app-key",
		Type:           "npm",
		Name:           "test-package",
		Versions:       []string{"1.0.0"},
	}

	mockPackageService := mockpackages.NewMockPackageService(ctrl)
	mockPackageService.EXPECT().BindPackage(gomock.Any(), requestPayload).
		Return(errors.New("bind error")).Times(1)

	cmd := &bindPackageCommand{
		packageService: mockPackageService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "bind error", err.Error())
}
