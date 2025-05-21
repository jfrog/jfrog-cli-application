package applications

import (
	"errors"
	"net/http"
	"testing"

	mockhttp "github.com/jfrog/jfrog-cli-application/apptrust/http/mocks"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	mockservice "github.com/jfrog/jfrog-cli-application/apptrust/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestApplicationService_CreateApplication(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *http.Response
		mockBody      []byte
		mockError     error
		expectedError string
	}{
		{
			name:          "CreateApplication successful",
			mockResponse:  &http.Response{StatusCode: http.StatusCreated},
			mockBody:      []byte(`{"application_key":"app-123"}`),
			mockError:     nil,
			expectedError: "",
		},
		{
			name:          "CreateApplication failed with non-201 status code",
			mockResponse:  &http.Response{StatusCode: http.StatusBadRequest},
			mockBody:      []byte(""),
			mockError:     nil,
			expectedError: "failed to create an application. Status code: 400.\n",
		},
		{
			name:          "CreateApplication failed with error",
			mockResponse:  nil,
			mockBody:      nil,
			mockError:     errors.New("http error"),
			expectedError: "http error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Post("/v1/applications", gomock.Any()).Return(tt.mockResponse, tt.mockBody, tt.mockError)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			as := NewApplicationService()
			err := as.CreateApplication(mockCtx, &model.AppDescriptor{ApplicationKey: "app-123"})

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
