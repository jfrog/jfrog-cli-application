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
			mockHttpClient.EXPECT().Post("/v1/applications/test-app/versions/", tt.request, nil).
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
		applicationKey   string
		version          string
		request          *model.UpdateAppVersionRequest
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectedError    string
	}{
		{
			name:           "success - tag only",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
			},
			mockResponse:     &http.Response{StatusCode: 202},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "success - properties only",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
			},
			mockResponse:     &http.Response{StatusCode: 202},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "success - delete properties only",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				DeleteProperties: []string{"legacy_param", "toBeDeleted"},
			},
			mockResponse:     &http.Response{StatusCode: 202},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "success - combined update",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
				Properties: map[string][]string{
					"status": {"rc", "validated"},
				},
				DeleteProperties: []string{"old_param"},
			},
			mockResponse:     &http.Response{StatusCode: 202},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:           "failure - 400",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				Tag: "invalid-tag",
			},
			mockResponse:     &http.Response{StatusCode: 400},
			mockResponseBody: "bad request",
			mockError:        nil,
			expectedError:    "failed to update app version",
		},
		{
			name:           "failure - 404",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
			},
			mockResponse:     &http.Response{StatusCode: 404},
			mockResponseBody: "not found",
			mockError:        nil,
			expectedError:    "failed to update app version",
		},
		{
			name:           "http client error",
			applicationKey: "test-app",
			version:        "1.0.0",
			request: &model.UpdateAppVersionRequest{
				Tag: "release/1.2.3",
			},
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectedError:    "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Patch("/v1/applications/test-app/versions/1.0.0", tt.request).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.UpdateAppVersion(mockCtx, tt.applicationKey, tt.version, tt.request)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
