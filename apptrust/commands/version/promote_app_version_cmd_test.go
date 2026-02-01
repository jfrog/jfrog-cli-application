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

func TestPromoteAppVersionCommand_Run(t *testing.T) {
	tests := []struct {
		name              string
		sync              bool
		overwriteStrategy string
	}{
		{
			name: "sync flag true",
			sync: true,
		},
		{
			name: "sync flag false",
			sync: false,
		},
		{
			name:              "with overwrite strategy disabled (sent as DISABLED)",
			sync:              true,
			overwriteStrategy: "DISABLED",
		},
		{
			name:              "with overwrite strategy latest (sent as LATEST)",
			sync:              true,
			overwriteStrategy: "LATEST",
		},
		{
			name:              "with overwrite strategy all (sent as ALL)",
			sync:              true,
			overwriteStrategy: "ALL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serverDetails := &config.ServerDetails{Url: "https://example.com"}
			applicationKey := "app-key"
			version := "1.0.0"
			requestPayload := &model.PromoteAppVersionRequest{
				Stage: "prod",
				CommonPromoteAppVersion: model.CommonPromoteAppVersion{
					OverwriteStrategy: tt.overwriteStrategy,
				},
			}

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			mockVersionService.EXPECT().PromoteAppVersion(gomock.Any(), applicationKey, version, requestPayload, tt.sync).
				Return(nil).Times(1)

			cmd := &promoteAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  serverDetails,
				applicationKey: applicationKey,
				version:        version,
				requestPayload: requestPayload,
				sync:           tt.sync,
			}

			err := cmd.Run()
			assert.NoError(t, err)
		})
	}
}

func TestPromoteAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"
	requestPayload := &model.PromoteAppVersionRequest{
		Stage: "prod",
	}
	sync := true
	expectedError := errors.New("service error occurred")

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().PromoteAppVersion(gomock.Any(), applicationKey, version, requestPayload, sync).
		Return(expectedError).Times(1)

	cmd := &promoteAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
		sync:           sync,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error occurred")
}
