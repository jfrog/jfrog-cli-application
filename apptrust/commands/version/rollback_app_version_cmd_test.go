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
	tests := []struct {
		name           string
		applicationKey string
		version        string
		fromStage      string
		sync           bool
		mockError      error
		expectedError  bool
	}{
		{
			name:           "successful rollback with sync=false",
			applicationKey: "video-encoder",
			version:        "1.5.0",
			fromStage:      "qa",
			sync:           false,
			mockError:      nil,
			expectedError:  false,
		},
		{
			name:           "successful rollback with sync=true",
			applicationKey: "test-app",
			version:        "1.0.0",
			fromStage:      "qa",
			sync:           true,
			mockError:      nil,
			expectedError:  false,
		},
		{
			name:           "failed rollback",
			applicationKey: "video-encoder",
			version:        "1.5.0",
			fromStage:      "qa",
			sync:           false,
			mockError:      errors.New("rollback service error occurred"),
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serverDetails := &config.ServerDetails{Url: "https://example.com"}
			requestPayload := &model.RollbackAppVersionRequest{
				FromStage: tt.fromStage,
			}

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			mockVersionService.EXPECT().RollbackAppVersion(gomock.Any(), tt.applicationKey, tt.version, requestPayload, tt.sync).
				Return(tt.mockError).Times(1)

			cmd := &rollbackAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  serverDetails,
				applicationKey: tt.applicationKey,
				version:        tt.version,
				requestPayload: requestPayload,
				fromStage:      tt.fromStage,
				sync:           tt.sync,
			}

			err := cmd.Run()

			if tt.expectedError {
				assert.Error(t, err)
				if tt.mockError != nil {
					assert.Contains(t, err.Error(), tt.mockError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
