package utils

import (
	"reflect"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/stretchr/testify/assert"
)

func TestParseSliceFlag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string{}},
		{"single value", "foo", []string{"foo"}},
		{"multiple values", "foo;bar;baz", []string{"foo", "bar", "baz"}},
		{"values with spaces", " foo ; bar ;baz ", []string{"foo", "bar", "baz"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSliceFlag(tt.input)
			assert.Equal(t, tt.expected, result, "ParseSliceFlag(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}

func TestParseMapFlag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string]string
		expectErr bool
	}{
		{"empty string", "", map[string]string{}, false},
		{"single pair", "foo=bar", map[string]string{"foo": "bar"}, false},
		{"multiple pairs", "foo=bar;baz=qux", map[string]string{"foo": "bar", "baz": "qux"}, false},
		{"pairs with spaces", " foo = bar ; baz = qux ", map[string]string{"foo": "bar", "baz": "qux"}, false},
		{"missing value", "foo=;bar=baz", map[string]string{"foo": "", "bar": "baz"}, false},
		{"missing key", "=bar", map[string]string{"": "bar"}, false},
		{"no equal sign", "foo;bar=baz", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMapFlag(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "ParseMapFlag(%q) expected error, got nil", tt.input)
				return
			}
			assert.NoError(t, err, "ParseMapFlag(%q) unexpected error: %v", tt.input, err)
			assert.True(t, reflect.DeepEqual(result, tt.expected), "ParseMapFlag(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}

func TestValidateEnumFlag(t *testing.T) {
	tests := []struct {
		name          string
		flagName      string
		value         string
		validValues   []string
		defaultValue  string
		expectError   bool
		expectedValue string
	}{
		{
			name:          "valid value",
			flagName:      "test-flag",
			value:         "foo",
			validValues:   []string{"foo", "bar", "baz"},
			defaultValue:  "",
			expectError:   false,
			expectedValue: "foo",
		},
		{
			name:          "invalid value with default",
			flagName:      "test-flag",
			value:         "invalid",
			validValues:   []string{"foo", "bar", "baz"},
			defaultValue:  "bar",
			expectError:   true,
			expectedValue: "",
		},
		{
			name:          "invalid value without default",
			flagName:      "test-flag",
			value:         "invalid",
			validValues:   []string{"foo", "bar", "baz"},
			defaultValue:  "",
			expectError:   true,
			expectedValue: "",
		},
		{
			name:          "empty value with default",
			flagName:      "test-flag",
			value:         "",
			validValues:   []string{"foo", "bar", "baz"},
			defaultValue:  "baz",
			expectError:   false,
			expectedValue: "baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateEnumFlag(tt.flagName, tt.value, tt.defaultValue, tt.validValues)
			if tt.expectError {
				assert.Error(t, err, "ValidateEnumFlag(%q) expected error, got nil", tt.value)
				return
			}
			assert.NoError(t, err, "ValidateEnumFlag(%q) unexpected error: %v", tt.value, err)
			assert.Equal(t, tt.expectedValue, result, "ValidateEnumFlag(%q) = %v, want %v", tt.value, result, tt.expectedValue)
		})
	}
}

func TestParseDelimitedSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected [][]string
	}{
		{"empty string", "", nil},
		{"single entry", "foo:bar", [][]string{{"foo", "bar"}}},
		{"multiple entries", "foo:bar;baz:qux", [][]string{{"foo", "bar"}, {"baz", "qux"}}},
		{"entries with extra parts", "a:1:2;b:3", [][]string{{"a", "1", "2"}, {"b", "3"}}},
		{"trailing separator", "foo:bar;", [][]string{{"foo", "bar"}, {""}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDelimitedSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseDelimitedSlice(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseNameVersionPairs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  [][2]string
		expectErr bool
	}{
		{"empty string", "", nil, false},
		{"single pair", "foo:1.0.0", [][2]string{{"foo", "1.0.0"}}, false},
		{"multiple pairs", "foo:1.0.0;bar:2.0.0", [][2]string{{"foo", "1.0.0"}, {"bar", "2.0.0"}}, false},
		{"spaces", " foo:1.0.0 ; bar:2.0.0 ", [][2]string{{" foo", "1.0.0 "}, {" bar", "2.0.0 "}}, false},
		{"invalid format", "foo", nil, true},
		{"too many parts", "foo:1.0.0:extra", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNameVersionPairs(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "expected error for input %q", tt.input)
				return
			}
			assert.NoError(t, err, "unexpected error for input %q: %v", tt.input, err)
			assert.Equal(t, tt.expected, result, "ParseNameVersionPairs(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}

func TestParseMonitorPolicyFlag(t *testing.T) {
	valueFive := 5
	valueThree := 3
	valueTwo := 2
	valueOne := 1

	tests := []struct {
		name      string
		input     string
		expected  *model.MonitorPolicy
		expectErr bool
		errorMsg  string
	}{
		{
			name:     "version count",
			input:    "type=version_count, value=5",
			expected: &model.MonitorPolicy{Type: model.MonitorPolicyTypeVersionCount, Value: &valueFive},
		},
		{
			name:     "time frame in months",
			input:    "type=time_frame_in_months, value=3",
			expected: &model.MonitorPolicy{Type: model.MonitorPolicyTypeTimeframe, Value: &valueThree},
		},
		{
			name:     "none without value",
			input:    "type=none",
			expected: &model.MonitorPolicy{Type: model.MonitorPolicyTypeNone},
		},
		{
			name:      "missing type",
			input:     "value=5",
			expectErr: true,
			errorMsg:  "missing required 'type' field",
		},
		{
			name:      "invalid type",
			input:     "type=bogus, value=5",
			expectErr: true,
			errorMsg:  "invalid type 'bogus'",
		},
		{
			name:      "missing value for enabled type",
			input:     "type=version_count",
			expectErr: true,
			errorMsg:  "'value' is required",
		},
		{
			name:      "value set for none",
			input:     "type=none, value=5",
			expectErr: true,
			errorMsg:  "'value' must not be set",
		},
		{
			name:      "non-integer value",
			input:     "type=version_count, value=abc",
			expectErr: true,
			errorMsg:  "must be an integer",
		},
		{
			name:      "zero value",
			input:     "type=version_count, value=0",
			expectErr: true,
			errorMsg:  "must be a positive integer",
		},
		{
			name:      "negative value",
			input:     "type=version_count, value=-1",
			expectErr: true,
			errorMsg:  "must be a positive integer",
		},
		{
			name:      "unknown field",
			input:     "type=version_count, value=5, foo=bar",
			expectErr: true,
			errorMsg:  "invalid field 'foo' (supported fields: type, value, minor_versions, patches_per_minor)",
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: true,
			errorMsg:  "missing required 'type' field",
		},
		{
			name:      "invalid key-value pair",
			input:     "type",
			expectErr: true,
			errorMsg:  "invalid key-value pair",
		},
		{
			name:     "minor patch with minor versions and patches per minor",
			input:    "type=version_count_minor_patch, minor_versions=2, patches_per_minor=1",
			expected: &model.MonitorPolicy{Type: model.MonitorPolicyTypeVersionCountMinorPatch, MinorVersions: &valueTwo, PatchesPerMinor: &valueOne},
		},
		{
			name:     "minor patch without patches per minor",
			input:    "type=version_count_minor_patch, minor_versions=3",
			expected: &model.MonitorPolicy{Type: model.MonitorPolicyTypeVersionCountMinorPatch, MinorVersions: &valueThree},
		},
		{
			name:      "minor patch without minor versions",
			input:     "type=version_count_minor_patch, patches_per_minor=1",
			expectErr: true,
			errorMsg:  "'minor_versions' is required when type is 'version_count_minor_patch'",
		},
		{
			name:      "minor patch with value",
			input:     "type=version_count_minor_patch, minor_versions=2, value=5",
			expectErr: true,
			errorMsg:  "'value' must not be set when type is 'version_count_minor_patch'",
		},
		{
			name:      "minor versions with version count",
			input:     "type=version_count, value=5, minor_versions=2",
			expectErr: true,
			errorMsg:  "'minor_versions' must not be set when type is 'version_count'",
		},
		{
			name:      "patches per minor with time frame",
			input:     "type=time_frame_in_months, value=3, patches_per_minor=1",
			expectErr: true,
			errorMsg:  "'patches_per_minor' must not be set when type is 'time_frame_in_months'",
		},
		{
			name:      "minor versions with none",
			input:     "type=none, minor_versions=2",
			expectErr: true,
			errorMsg:  "'minor_versions' must not be set when type is 'none'",
		},
		{
			name:      "zero minor versions",
			input:     "type=version_count_minor_patch, minor_versions=0",
			expectErr: true,
			errorMsg:  "'minor_versions' must be a positive integer",
		},
		{
			name:      "negative patches per minor",
			input:     "type=version_count_minor_patch, minor_versions=2, patches_per_minor=-1",
			expectErr: true,
			errorMsg:  "'patches_per_minor' must be a positive integer",
		},
		{
			name:      "non-integer minor versions",
			input:     "type=version_count_minor_patch, minor_versions=two",
			expectErr: true,
			errorMsg:  "'minor_versions' must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMonitorPolicyFlag(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "ParseMonitorPolicyFlag(%q) expected error, got nil", tt.input)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "error message should contain %q", tt.errorMsg)
				}
				return
			}
			assert.NoError(t, err, "ParseMonitorPolicyFlag(%q) unexpected error: %v", tt.input, err)
			assert.Equal(t, tt.expected, result, "ParseMonitorPolicyFlag(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}

func TestParseLabelKeyValuePairs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []model.LabelEntry
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "empty string",
			input:     "",
			expected:  []model.LabelEntry{},
			expectErr: false,
		},
		{
			name:      "single pair",
			input:     "key1=value1",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value1"}},
			expectErr: false,
		},
		{
			name:      "multiple pairs",
			input:     "key1=value1;key2=value2;key3=value3",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}, {Key: "key3", Value: "value3"}},
			expectErr: false,
		},
		{
			name:      "same key multiple values",
			input:     "environment=production;environment=staging;region=us-east",
			expected:  []model.LabelEntry{{Key: "environment", Value: "production"}, {Key: "environment", Value: "staging"}, {Key: "region", Value: "us-east"}},
			expectErr: false,
		},
		{
			name:      "whitespace handling",
			input:     " key1 = value1 ; key2 = value2 ",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "empty pairs skipped",
			input:     "key1=value1;;key2=value2",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "leading and trailing separators",
			input:     ";key1=value1;key2=value2;",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "empty value",
			input:     "key1=;key2=value2",
			expected:  []model.LabelEntry{{Key: "key1", Value: ""}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "value with equals sign",
			input:     "key1=value=with=equals;key2=normal",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value=with=equals"}, {Key: "key2", Value: "normal"}},
			expectErr: false,
		},
		{
			name:      "missing equals sign",
			input:     "key1=value1;invalid-format;key2=value2",
			expectErr: true,
			errorMsg:  "invalid key-value pair",
		},
		{
			name:      "no equals sign at all",
			input:     "invalid-format",
			expectErr: true,
			errorMsg:  "invalid key-value pair",
		},
		{
			name:      "empty key",
			input:     "=value1;key2=value2",
			expected:  []model.LabelEntry{{Key: "", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "whitespace only pairs skipped",
			input:     "key1=value1;   ;key2=value2",
			expected:  []model.LabelEntry{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "special characters in key and value",
			input:     "env-name=prod-env;region=us-east-1",
			expected:  []model.LabelEntry{{Key: "env-name", Value: "prod-env"}, {Key: "region", Value: "us-east-1"}},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLabelKeyValuePairs(tt.input)
			if tt.expectErr {
				assert.Error(t, err, "ParseLabelKeyValuePairs(%q) expected error, got nil", tt.input)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "error message should contain %q", tt.errorMsg)
				}
				return
			}
			assert.NoError(t, err, "ParseLabelKeyValuePairs(%q) unexpected error: %v", tt.input, err)
			assert.Equal(t, tt.expected, result, "ParseLabelKeyValuePairs(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}
