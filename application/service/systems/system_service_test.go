package systems

import (
	"errors"
	"net/http"
	"testing"

	mockservice "github.com/jfrog/jfrog-cli-application/application/service/mocks"
	"go.uber.org/mock/gomock"

	mockhttp "github.com/jfrog/jfrog-cli-application/application/http/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSystemService_Ping(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *http.Response
		mockBody      []byte
		mockError     error
		expectedError error
	}{
		{
			name: "Ping successful",
			mockResponse: &http.Response{
				StatusCode: 200,
			},
			mockBody:      []byte("pong"),
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "Ping failed with non-200 status code",
			mockResponse: &http.Response{
				StatusCode: 500,
			},
			mockBody:      []byte(""),
			mockError:     nil,
			expectedError: errors.New("failed pinging application service. Status code: 500"),
		},
		{
			name:          "Ping failed with error",
			mockResponse:  nil,
			mockBody:      nil,
			mockError:     errors.New("http error"),
			expectedError: errors.New("http error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHttpClient := mockhttp.NewMockAppHttpClient(ctrl)
			mockHttpClient.EXPECT().Get("/v1/system/ping").
				Return(tt.mockResponse, tt.mockBody, tt.mockError)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			ss := NewSystemService()
			err := ss.Ping(mockCtx)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
