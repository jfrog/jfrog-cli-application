package http

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"encoding/json"
	"fmt"
	"net/http"

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

type ApptrustHttpClient interface {
	GetHttpClient() *jfroghttpclient.JfrogHttpClient
	Post(path string, requestBody interface{}, params map[string]string) (resp *http.Response, body []byte, err error)
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
