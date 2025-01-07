package utils

import (
	"fmt"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

func AssertValueProvided(c *components.Context, fieldName string) error {
	if c.GetStringFlagValue(fieldName) == "" {
		return errorutils.CheckErrorf("the --%s option is mandatory", fieldName)
	}
	return nil
}

func ServerDetailsByFlags(ctx *components.Context) (*coreConfig.ServerDetails, error) {
	serverDetails, err := pluginsCommon.CreateServerDetailsWithConfigOffer(ctx, true, commonCliUtils.Platform)
	if err != nil {
		return nil, err
	}
	if serverDetails.Url == "" {
		return nil, fmt.Errorf("platform URL is mandatory for evidence commands")
	}
	if serverDetails.GetUser() != "" && serverDetails.GetPassword() != "" {
		return nil, fmt.Errorf("evidence service does not support basic authentication")
	}

	return serverDetails, nil
}
