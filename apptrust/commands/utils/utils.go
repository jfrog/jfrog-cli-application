package utils

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
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
		return nil, fmt.Errorf("platform URL is mandatory for AppTrust commands")
	}
	if serverDetails.GetUser() != "" && serverDetails.GetPassword() != "" {
		return nil, fmt.Errorf("AppTrust service does not support basic authentication")
	}

	return serverDetails, nil
}

// ParseSliceFlag parses a semicolon-separated string into a slice of strings.
func ParseSliceFlag(flagValue string) []string {
	if flagValue == "" {
		return []string{}
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
	return ParseKeyValueString(flagValue, ";")
}

func ParseKeyValueString(value, separator string) (map[string]string, error) {
	result := make(map[string]string)
	if value == "" {
		return result, nil
	}
	pairs := strings.Split(value, separator)
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

func ParseMonitorPolicyFlag(flagValue string) (*model.MonitorPolicy, error) {
	fields, err := ParseKeyValueString(flagValue, ",")
	if err != nil {
		return nil, err
	}

	numberFields := []string{"value", "minor_versions", "patches_per_minor"}
	for key := range fields {
		if key != "type" && !slices.Contains(numberFields, key) {
			return nil, errorutils.CheckErrorf("invalid field '%s' (supported fields: type, %s)", key, strings.Join(numberFields, ", "))
		}
	}

	policyType, typeSet := fields["type"]
	if !typeSet {
		return nil, errorutils.CheckErrorf("missing required 'type' field")
	}
	if !slices.Contains(model.MonitorPolicyTypeValues, policyType) {
		return nil, errorutils.CheckErrorf("invalid type '%s' (supported types: %s)", policyType, coreutils.ListToText(model.MonitorPolicyTypeValues))
	}

	var required, allowed []string
	switch policyType {
	case model.MonitorPolicyTypeTimeframe, model.MonitorPolicyTypeVersionCount:
		required, allowed = []string{"value"}, []string{"value"}
	case model.MonitorPolicyTypeVersionCountMinorPatch:
		required, allowed = []string{"minor_versions"}, []string{"minor_versions", "patches_per_minor"}
	}
	for _, key := range numberFields {
		_, set := fields[key]
		if set && !slices.Contains(allowed, key) {
			return nil, errorutils.CheckErrorf("'%s' must not be set when type is '%s'", key, policyType)
		}
		if !set && slices.Contains(required, key) {
			return nil, errorutils.CheckErrorf("'%s' is required when type is '%s'", key, policyType)
		}
	}

	policy := &model.MonitorPolicy{Type: policyType}
	if policy.Value, err = parsePositiveIntField(fields, "value"); err != nil {
		return nil, err
	}
	if policy.MinorVersions, err = parsePositiveIntField(fields, "minor_versions"); err != nil {
		return nil, err
	}
	if policy.PatchesPerMinor, err = parsePositiveIntField(fields, "patches_per_minor"); err != nil {
		return nil, err
	}
	return policy, nil
}

// parsePositiveIntField returns nil when the field is not set.
func parsePositiveIntField(fields map[string]string, key string) (*int, error) {
	valueStr, set := fields[key]
	if !set {
		return nil, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return nil, errorutils.CheckErrorf("'%s' must be an integer: %s", key, err.Error())
	}
	if value <= 0 {
		return nil, errorutils.CheckErrorf("'%s' must be a positive integer", key)
	}
	return &value, nil
}

func ParseLabelKeyValuePairs(flagValue string) ([]model.LabelEntry, error) {
	if flagValue == "" {
		return []model.LabelEntry{}, nil
	}

	var result []model.LabelEntry
	pairs := strings.Split(flagValue, ";")
	for _, pair := range pairs {
		trimmedPair := strings.TrimSpace(pair)
		if trimmedPair == "" {
			continue
		}
		keyValue := strings.SplitN(trimmedPair, "=", 2)
		if len(keyValue) != 2 {
			return nil, errorutils.CheckErrorf("invalid key-value pair: '%s' (expected format key=value)", pair)
		}
		result = append(result, model.LabelEntry{
			Key:   strings.TrimSpace(keyValue[0]),
			Value: strings.TrimSpace(keyValue[1]),
		})
	}
	return result, nil
}
