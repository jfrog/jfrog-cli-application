package version

import (
	"errors"
	"testing"

	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestRollbackAppVersionCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "video-encoder"
	version := "1.5.0"
	requestPayload := &model.RollbackAppVersionRequest{
		FromStage: "qa",
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().RollbackAppVersion(gomock.Any(), applicationKey, version, requestPayload, false).
		Return(nil).Times(1)

	cmd := &rollbackAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestRollbackAppVersionCommand_Run_WithSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applicationKey := "test-app"
	version := "1.0.0"
	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.RollbackAppVersionRequest{
		FromStage: "qa",
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().RollbackAppVersion(gomock.Any(), applicationKey, version, requestPayload, true).
		Return(nil).Times(1)

	cmd := &rollbackAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
		fromStage:      "qa",
		sync:           true,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestRollbackAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "video-encoder"
	version := "1.5.0"
	requestPayload := &model.RollbackAppVersionRequest{
		FromStage: "qa",
	}
	expectedError := errors.New("rollback service error occurred")

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().RollbackAppVersion(gomock.Any(), applicationKey, version, requestPayload, false).
		Return(expectedError).Times(1)

	cmd := &rollbackAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback service error occurred")
}
