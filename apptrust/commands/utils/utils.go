package utils

import (
	"fmt"
	"slices"
	"strings"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	pluginsCommon "github.com/jfrog/jfrog-cli-core/v2/plugins/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const (
	EntrySeparator = ";"
	PartSeparator  = ":"
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

// ValidateEnumFlag validates that a flag value is in the list of allowed values.
// If the value is empty, returns the default value.
// Otherwise, returns an error if the value is not in the allowed values.
func ValidateEnumFlag(flagName, value string, defaultValue string, allowedValues []string) (string, error) {
	if value == "" {
		return defaultValue, nil
	}

	if slices.Contains(allowedValues, value) {
		return value, nil
	}

	return "", errorutils.CheckErrorf("invalid value for --%s: '%s'. Allowed values: %s",
		flagName, value, coreutils.ListToText(allowedValues))
}

// ParsePackagesFlag parses a comma-separated list of package name:version pairs into a slice of maps.
// Each map contains keys "name" and "version". Returns an error if any entry is not in the expected format.
// Example input: "pkg1:1.0.0,pkg2:2.0.0" => []map[string]string{{"name": "pkg1", "version": "1.0.0"}, {"name": "pkg2", "version": "2.0.0"}}
func ParsePackagesFlag(flagValue string) ([]map[string]string, error) {
	if flagValue == "" {
		return nil, nil
	}
	pairs := strings.Split(flagValue, ",")
	var result []map[string]string
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), PartSeparator, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid package format: %s (expected <name>:<version>)", pair)
		}
		result = append(result, map[string]string{"name": parts[0], "version": parts[1]})
	}
	return result, nil
}

// ParseDelimitedSlice splits a delimited string into a slice of string slices.
// Example: input "a:1;b:2" returns [][]string{{"a","1"},{"b","2"}}
func ParseDelimitedSlice(input string) [][]string {
	var result [][]string
	if input == "" {
		return result
	}
	entries := strings.Split(input, EntrySeparator)
	for _, entry := range entries {
		parts := strings.Split(entry, PartSeparator)
		result = append(result, parts)
	}
	return result
}

// ParseNameVersionPairs parses a delimited string (e.g., "name1:version1;name2:version2") into a slice of [2]string pairs.
// Returns an error if any entry does not have exactly two parts.
func ParseNameVersionPairs(input string) ([][2]string, error) {
	var result [][2]string
	if input == "" {
		return result, nil
	}
	for _, parts := range ParseDelimitedSlice(input) {
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %v", parts)
		}
		result = append(result, [2]string{parts[0], parts[1]})
	}
	return result, nil
}

// ParseListPropertiesFlag parses a properties string into a map of keys to value slices.
// Format: "key1=value1[,value2,...];key2=value3[,value4,...]"
// Examples:
//   - "status=rc" -> {"status": ["rc"]}
//   - "status=rc,validated" -> {"status": ["rc", "validated"]}
//   - "status=rc;deployed_to=staging" -> {"status": ["rc"], "deployed_to": ["staging"]}
//   - "old_flag=" -> {"old_flag": []} (clears values)
func ParseListPropertiesFlag(propertiesStr string) (map[string][]string, error) {
	if propertiesStr == "" {
		return nil, nil
	}

	result := make(map[string][]string)
	pairs := strings.Split(propertiesStr, ";")

	for _, pair := range pairs {
		keyValue := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(keyValue) != 2 {
			return nil, errorutils.CheckErrorf("invalid property format: \"%s\" (expected key=value1[,value2,...])", pair)
		}

		key := strings.TrimSpace(keyValue[0])
		valuesStr := strings.TrimSpace(keyValue[1])

		if key == "" {
			return nil, errorutils.CheckErrorf("property key cannot be empty")
		}

		var values []string
		if valuesStr != "" {
			values = strings.Split(valuesStr, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
		} else {
			// Return empty slice instead of nil for empty values
			values = []string{}
		}
		// Always set the key, even with empty values (to clear values)
		result[key] = values
	}

	return result, nil
}
