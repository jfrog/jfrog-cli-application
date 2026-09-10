package http

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/jfrog/jfrog-client-go/utils/log"

	commonCliConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
)

const apptrustApiPath = "apptrust/api"

type MultipartPart struct {
	Name        string
	Filename    string
	ContentType string
	Body        []byte
	Path        string
}

type ApptrustHttpClient interface {
	GetHttpClient() *jfroghttpclient.JfrogHttpClient
	Post(path string, requestBody interface{}, params map[string]string) (resp *http.Response, body []byte, err error)
	PostMultipart(path string, parts []MultipartPart, params map[string]string) (resp *http.Response, body []byte, err error)
	Get(path string, params map[string]string) (resp *http.Response, body []byte, err error)
	Patch(path string, requestBody interface{}, params map[string]string) (resp *http.Response, body []byte, err error)
	Delete(path string, params map[string]string) (resp *http.Response, body []byte, err error)
}

type apptrustHttpClient struct {
	client        *jfroghttpclient.JfrogHttpClient
	serverDetails *commonCliConfig.ServerDetails
	authDetails   auth.ServiceDetails
	serviceConfig clientConfig.Config
}

func NewAppHttpClient(serverDetails *commonCliConfig.ServerDetails) (ApptrustHttpClient, error) {
	certsPath, err := coreutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}

	authDetails, err := serverDetails.CreateLifecycleAuthConfig()
	if err != nil {
		return nil, err
	}

	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(authDetails).
		SetCertificatesPath(certsPath).
		SetInsecureTls(serverDetails.InsecureTls).
		SetHttpRetries(1).
		Build()
	if err != nil {
		return nil, err
	}

	jfHttpClient, err := jfroghttpclient.JfrogClientBuilder().
		SetCertificatesPath(certsPath).
		SetInsecureTls(serviceConfig.IsInsecureTls()).
		SetClientCertPath(serverDetails.GetClientCertPath()).
		SetClientCertKeyPath(serverDetails.GetClientCertKeyPath()).
		AppendPreRequestInterceptor(authDetails.RunPreRequestFunctions).
		SetContext(serviceConfig.GetContext()).
		SetDialTimeout(serviceConfig.GetDialTimeout()).
		SetOverallRequestTimeout(serviceConfig.GetOverallRequestTimeout()).
		SetRetries(serviceConfig.GetHttpRetries()).
		SetRetryWaitMilliSecs(serviceConfig.GetHttpRetryWaitMilliSecs()).
		Build()
	if err != nil {
		return nil, err
	}

	appClient := &apptrustHttpClient{
		client:        jfHttpClient,
		serverDetails: serverDetails,
		authDetails:   authDetails,
		serviceConfig: serviceConfig,
	}
	return appClient, nil
}

func (c *apptrustHttpClient) GetHttpClient() *jfroghttpclient.JfrogHttpClient {
	return c.client
}

func (c *apptrustHttpClient) Post(path string, requestBody interface{}, params map[string]string) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, apptrustApiPath+path, params)
	if err != nil {
		return nil, nil, err
	}

	requestContent, err := c.toJsonBytes(requestBody)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Sending POST request to:", url)
	return c.client.SendPost(url, requestContent, c.getJsonHttpClientDetails())
}

func (c *apptrustHttpClient) PostMultipart(path string, parts []MultipartPart, params map[string]string) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, apptrustApiPath+path, params)
	if err != nil {
		return nil, nil, err
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	contentType := mw.FormDataContentType()
	writeErrCh := make(chan error, 1)
	go func() {
		writeErr := writeMultipartParts(mw, parts)
		_ = pw.CloseWithError(writeErr)
		writeErrCh <- writeErr
	}()

	log.Debug("Sending multipart POST request to:", url)
	resp, body, err = c.sendMultipartPost(url, pr, contentType)
	_ = pr.Close()
	if writeErr := <-writeErrCh; writeErr != nil {
		return nil, nil, writeErr
	}
	return resp, body, err
}

func writeMultipartParts(mw *multipart.Writer, parts []MultipartPart) error {
	for _, part := range parts {
		if err := writeMultipartPart(mw, part); err != nil {
			return err
		}
	}
	return mw.Close()
}

func writeMultipartPart(mw *multipart.Writer, part MultipartPart) (err error) {
	filename := part.Filename
	if filename == "" && part.Path != "" {
		filename = filepath.Base(part.Path)
	}

	header := make(textproto.MIMEHeader)
	filenamePart := ""
	if filename != "" {
		filenamePart = fmt.Sprintf(`; filename="%s"`, filename)
	}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q%s`, part.Name, filenamePart))
	if part.ContentType != "" {
		header.Set("Content-Type", part.ContentType)
	}
	writer, err := mw.CreatePart(header)
	if err != nil {
		return err
	}
	if part.Path == "" {
		_, err = writer.Write(part.Body)
		return err
	}

	// File part
	file, err := os.Open(part.Path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = io.Copy(writer, file)
	return err
}

func (c *apptrustHttpClient) sendMultipartPost(url string, body io.Reader, contentType string) (*http.Response, []byte, error) {
	details := c.authDetails.CreateHttpClientDetails()
	details.AddHeader("Content-Type", contentType)
	return c.client.SendPostFromReader(url, body, &details)
}

func (c *apptrustHttpClient) Get(path string, params map[string]string) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, apptrustApiPath+path, params)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Sending GET request to:", url)
	response, body, _, err := c.client.SendGet(url, false, c.getJsonHttpClientDetails())
	return response, body, err
}

func (c *apptrustHttpClient) Patch(path string, requestBody interface{}, params map[string]string) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, apptrustApiPath+path, params)
	if err != nil {
		return nil, nil, err
	}

	requestContent, err := c.toJsonBytes(requestBody)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Sending PATCH request to:", url)
	return c.client.SendPatch(url, requestContent, c.getJsonHttpClientDetails())
}

func (c *apptrustHttpClient) toJsonBytes(payload interface{}) ([]byte, error) {
	if payload == nil {
		return nil, fmt.Errorf("request payload is required")
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return jsonBytes, nil
}

func (c *apptrustHttpClient) Delete(path string, params map[string]string) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, apptrustApiPath+path, params)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Sending DELETE request to:", url)
	return c.client.SendDelete(url, nil, c.getJsonHttpClientDetails())
}

func (c *apptrustHttpClient) getJsonHttpClientDetails() *httputils.HttpClientDetails {
	httpClientDetails := c.authDetails.CreateHttpClientDetails()
	httpClientDetails.SetContentTypeApplicationJson()
	return &httpClientDetails
}
