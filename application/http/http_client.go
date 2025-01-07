package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	commonCliConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
)

const applicationApiPath = "application/api"

type AppHttpClient interface {
	GetHttpClient() *jfroghttpclient.JfrogHttpClient
	Post(path string, requestBody interface{}) (resp *http.Response, body []byte, err error)
	Get(path string) (resp *http.Response, body []byte, err error)
}

type appHttpClient struct {
	client        *jfroghttpclient.JfrogHttpClient
	serverDetails *commonCliConfig.ServerDetails
	authDetails   auth.ServiceDetails
	serviceConfig clientConfig.Config
}

func NewAppHttpClient(serverDetails *commonCliConfig.ServerDetails) (AppHttpClient, error) {
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

	appClient := &appHttpClient{
		client:        jfHttpClient,
		serverDetails: serverDetails,
		authDetails:   authDetails,
		serviceConfig: serviceConfig,
	}
	return appClient, nil
}

func (c *appHttpClient) GetHttpClient() *jfroghttpclient.JfrogHttpClient {
	return c.client
}

func (c *appHttpClient) Post(path string, requestBody interface{}) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, applicationApiPath+path, nil)
	if err != nil {
		return nil, nil, err
	}

	requestContent, err := c.toJsonBytes(requestBody)
	if err != nil {
		return nil, nil, err
	}

	println("url: ", url)
	return c.client.SendPost(url, requestContent, c.getJsonHttpClientDetails())
}

func (c *appHttpClient) Get(path string) (resp *http.Response, body []byte, err error) {
	url, err := utils.BuildUrl(c.serverDetails.Url, applicationApiPath+path, nil)
	if err != nil {
		return nil, nil, err
	}

	response, body, _, err := c.client.SendGet(url, false, c.getJsonHttpClientDetails())
	return response, body, err
}

func (c *appHttpClient) toJsonBytes(payload interface{}) ([]byte, error) {
	if payload == nil {
		return nil, fmt.Errorf("request payload is required")
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return jsonBytes, nil
}

func (c *appHttpClient) getJsonHttpClientDetails() *httputils.HttpClientDetails {
	httpClientDetails := c.authDetails.CreateHttpClientDetails()
	httpClientDetails.SetContentTypeApplicationJson()
	return &httpClientDetails
}
