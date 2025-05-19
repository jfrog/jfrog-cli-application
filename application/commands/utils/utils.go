package utils

import (
	"fmt"
	"strings"

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

// ParseSliceFlag parses a comma-separated string into a slice of strings.
func ParseSliceFlag(flagValue string) []string {
	if flagValue == "" {
		return nil
	}
	values := strings.Split(flagValue, ";")

	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}
	return values
}

// ParseMapFlag parses a semicolon-separated string of key=value pairs into a map[string]string.
// Returns an error if any pair does not contain exactly one '='.
func ParseMapFlag(flagValue string) (map[string]string, error) {
	if flagValue == "" {
		return nil, nil
	}
	result := make(map[string]string)
	pairs := strings.Split(flagValue, ";")
	for _, pair := range pairs {
		keyValue := strings.SplitN(pair, "=", 2)
		if len(keyValue) != 2 {
			return nil, errorutils.CheckErrorf("invalid key-value pair: '%s' (expected format key=value)", pair)
		}
		result[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}
	return result, nil
}
