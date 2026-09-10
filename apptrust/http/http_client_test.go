package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	commonCliConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostMultipart(t *testing.T) {
	fileContent := []byte("PK zip-bytes")
	fieldBody := []byte(`{"mode":"path_mapping"}`)
	fileName := "payload.zip"

	var gotField, gotFile []byte
	var gotFilename, gotFieldContentType, gotFileContentType string
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		reader, err := r.MultipartReader()
		require.NoError(t, err)

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			body, err := io.ReadAll(part)
			require.NoError(t, err)

			switch part.FormName() {
			case "metadata":
				gotField = body
				gotFieldContentType = part.Header.Get("Content-Type")
			case "payload":
				gotFile = body
				gotFilename = part.FileName()
				gotFileContentType = part.Header.Get("Content-Type")
			default:
				t.Errorf("unexpected multipart part %q", part.FormName())
			}
		}

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	filePath := filepath.Join(t.TempDir(), fileName)
	require.NoError(t, os.WriteFile(filePath, fileContent, 0o600))

	client, err := NewAppHttpClient(&commonCliConfig.ServerDetails{Url: server.URL + "/"})
	require.NoError(t, err)

	parts := []MultipartPart{
		{Name: "metadata", ContentType: "application/json", Body: fieldBody},
		{Name: "payload", ContentType: "application/zip", Path: filePath},
	}
	resp, body, err := client.PostMultipart("/v1/resources/upload", parts, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, `{"ok":true}`, string(body))
	assert.Equal(t, "/apptrust/api/v1/resources/upload", gotPath)
	assert.Equal(t, fieldBody, gotField)
	assert.Equal(t, "application/json", gotFieldContentType)
	assert.Equal(t, fileContent, gotFile)
	assert.Equal(t, fileName, gotFilename)
	assert.Equal(t, "application/zip", gotFileContentType)
}
