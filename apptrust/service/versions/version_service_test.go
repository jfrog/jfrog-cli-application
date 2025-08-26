package versions

import (
	"errors"
	"net/http"
	"strconv"
	"testing"

	mockhttp "github.com/jfrog/jfrog-cli-application/apptrust/http/mocks"
	mockservice "github.com/jfrog/jfrog-cli-application/apptrust/service/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewVersionService()

	tests := []struct {
		name             string
		request          *model.CreateAppVersionRequest
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectedError    string
	}{
		{
			name:             "success",
			request:          &model.CreateAppVersionRequest{ApplicationKey: "test-app", Version: "1.0.0"},
			mockResponse:     &http.Response{StatusCode: 201},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:             "failure",
			request:          &model.CreateAppVersionRequest{ApplicationKey: "test-app", Version: "1.0.0"},
			mockResponse:     &http.Response{StatusCode: 400},
			mockResponseBody: "error",
			mockError:        nil,
			expectedError:    "failed to create app version",
		},
		{
			name:             "http client error",
			request:          &model.CreateAppVersionRequest{ApplicationKey: "test-app", Version: "1.0.0"},
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectedError:    "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Post("/v1/applications/test-app/versions/", tt.request, map[string]string{"async": "false"}).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.CreateAppVersion(mockCtx, tt.request)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestPromoteAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewVersionService()

	tests := []struct {
		name             string
		applicationKey   string
		version          string
		payload          *model.PromoteAppVersionRequest
		sync             bool
		expectedEndpoint string
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectedError    string
	}{
		{
			name:           "success with sync=true",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: &model.PromoteAppVersionRequest{
				Stage: "prod",
				CommonPromoteAppVersion: model.CommonPromoteAppVersion{
					PromotionType:          model.PromotionTypeCopy,
					IncludedRepositoryKeys: []string{"repo1", "repo2"},
					ExcludedRepositoryKeys: []string{"repo3"},
				},
			},
			sync:             true,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/promote",
			mockResponse:     &http.Response{StatusCode: 200},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "success with sync=false",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: &model.PromoteAppVersionRequest{
				Stage: "prod",
				CommonPromoteAppVersion: model.CommonPromoteAppVersion{
					PromotionType:          model.PromotionTypeCopy,
					IncludedRepositoryKeys: []string{"repo1", "repo2"},
					ExcludedRepositoryKeys: []string{"repo3"},
				},
			},
			sync:             false,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/promote",
			mockResponse:     &http.Response{StatusCode: 202},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "failure",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: &model.PromoteAppVersionRequest{
				Stage: "prod",
				CommonPromoteAppVersion: model.CommonPromoteAppVersion{
					PromotionType: model.PromotionTypeCopy,
				},
			},
			sync:             true,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/promote",
			mockResponse:     &http.Response{StatusCode: 400},
			mockResponseBody: "error",
			mockError:        nil,
			expectedError:    "failed to promote app version",
		},
		{
			name:           "http client error",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: &model.PromoteAppVersionRequest{
				Stage: "prod",
				CommonPromoteAppVersion: model.CommonPromoteAppVersion{
					PromotionType: model.PromotionTypeCopy,
				},
			},
			sync:             false,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/promote",
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectedError:    "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Post(tt.expectedEndpoint, tt.payload, map[string]string{"async": strconv.FormatBool(!tt.sync)}).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.PromoteAppVersion(mockCtx, tt.applicationKey, tt.version, tt.payload, tt.sync)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestReleaseAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewVersionService()

	tests := []struct {
		name             string
		applicationKey   string
		version          string
		payload          *model.ReleaseAppVersionRequest
		sync             bool
		expectedEndpoint string
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectedError    string
	}{
		{
			name:           "success with sync=true",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: model.NewReleaseAppVersionRequest(
				model.PromotionTypeCopy,
				[]string{"repo1", "repo2"},
				[]string{"repo3"},
				map[string]string{"key1": "value1"},
			),
			sync:             true,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/release",
			mockResponse:     &http.Response{StatusCode: 200},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "success with sync=false",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: model.NewReleaseAppVersionRequest(
				model.PromotionTypeCopy,
				[]string{"repo1", "repo2"},
				[]string{"repo3"},
				map[string]string{"key1": "value1"},
			),
			sync:             false,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/release",
			mockResponse:     &http.Response{StatusCode: 202},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "failure",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: model.NewReleaseAppVersionRequest(
				model.PromotionTypeCopy,
				nil,
				nil,
				nil,
			),
			sync:             true,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/release",
			mockResponse:     &http.Response{StatusCode: 400},
			mockResponseBody: "error",
			mockError:        nil,
			expectedError:    "failed to release app version",
		},
		{
			name:           "http client error",
			applicationKey: "test-app",
			version:        "1.0.0",
			payload: model.NewReleaseAppVersionRequest(
				model.PromotionTypeCopy,
				nil,
				nil,
				nil,
			),
			sync:             false,
			expectedEndpoint: "/v1/applications/test-app/versions/1.0.0/release",
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectedError:    "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Post(tt.expectedEndpoint, tt.payload, map[string]string{"async": strconv.FormatBool(!tt.sync)}).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.ReleaseAppVersion(mockCtx, tt.applicationKey, tt.version, tt.payload, tt.sync)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestUpdateAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewVersionService()

	tests := []struct {
		name             string
		request          *model.UpdateAppVersionRequest
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectError      bool
		errorMsg         string
	}{
		{
			name: "success - tag only",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
			},
			mockResponse:     &http.Response{StatusCode: http.StatusOK},
			mockResponseBody: "{}",
			mockError:        nil,
			expectError:      false,
			errorMsg:         "",
		},
		{
			name: "success - properties only",
			request: &model.UpdateAppVersionRequest{
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
			},
			mockResponse:     &http.Response{StatusCode: http.StatusOK},
			mockResponseBody: "{}",
			mockError:        nil,
			expectError:      false,
			errorMsg:         "",
		},
		{
			name: "success - delete properties only",
			request: &model.UpdateAppVersionRequest{
				DeleteProperties: []string{"legacy_param", "toBeDeleted"},
			},
			mockResponse:     &http.Response{StatusCode: http.StatusOK},
			mockResponseBody: "{}",
			mockError:        nil,
			expectError:      false,
			errorMsg:         "",
		},
		{
			name: "success - combined update",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
				DeleteProperties: []string{"old_param"},
			},
			mockResponse:     &http.Response{StatusCode: http.StatusOK},
			mockResponseBody: "{}",
			mockError:        nil,
			expectError:      false,
			errorMsg:         "",
		},
		{
			name: "failure - 400",
			request: &model.UpdateAppVersionRequest{
				Tag: "invalid-tag",
			},
			mockResponse:     &http.Response{StatusCode: http.StatusBadRequest},
			mockResponseBody: "bad request",
			mockError:        nil,
			expectError:      true,
			errorMsg:         "failed to update app version",
		},
		{
			name: "failure - 404",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
			},
			mockResponse:     &http.Response{StatusCode: http.StatusNotFound},
			mockResponseBody: "not found",
			mockError:        nil,
			expectError:      true,
			errorMsg:         "failed to update app version",
		},
		{
			name: "http client error",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
			},
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectError:      true,
			errorMsg:         "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Patch("/v1/applications/test-app/versions/1.0.0", tt.request).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).AnyTimes()

			err := service.UpdateAppVersion(mockCtx, "test-app", "1.0.0", tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRollbackAppVersion(t *testing.T) {
	tests := []struct {
		name           string
		applicationKey string
		version        string
		payload        *model.RollbackAppVersionRequest
		sync           bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "successful rollback with sync=true",
			applicationKey: "video-encoder",
			version:        "1.5.0",
			payload: &model.RollbackAppVersionRequest{
				FromStage: "qa",
			},
			sync:           true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "successful rollback with sync=false",
			applicationKey: "video-encoder",
			version:        "1.5.0",
			payload: &model.RollbackAppVersionRequest{
				FromStage: "prod",
			},
			sync:           false,
			expectedStatus: http.StatusAccepted,
			expectedError:  false,
		},
		{
			name:           "failed rollback - bad request",
			applicationKey: "invalid-app",
			version:        "1.0.0",
			payload: &model.RollbackAppVersionRequest{
				FromStage: "nonexistent",
			},
			sync:           true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "failed rollback - sync=true but got 202",
			applicationKey: "video-encoder",
			version:        "1.5.0",
			payload: &model.RollbackAppVersionRequest{
				FromStage: "qa",
			},
			sync:           true,
			expectedStatus: http.StatusAccepted,
			expectedError:  true,
		},
		{
			name:           "failed rollback - sync=false but got 200",
			applicationKey: "video-encoder",
			version:        "1.5.0",
			payload: &model.RollbackAppVersionRequest{
				FromStage: "prod",
			},
			sync:           false,
			expectedStatus: http.StatusOK,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCtx := mockservice.NewMockContext(ctrl)
			mockClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockClient)

			expectedEndpoint := "/v1/applications/" + tt.applicationKey + "/versions/" + tt.version + "/rollback"
			mockClient.EXPECT().Post(expectedEndpoint, tt.payload, map[string]string{"async": strconv.FormatBool(!tt.sync)}).
				Return(&http.Response{StatusCode: tt.expectedStatus}, []byte(""), nil)

			service := NewVersionService()
			err := service.RollbackAppVersion(mockCtx, tt.applicationKey, tt.version, tt.payload, tt.sync)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to rollback app version")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
