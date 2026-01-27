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

func TestParseLabelKeyValuePairs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []model.LabelKeyValue
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "empty string",
			input:     "",
			expected:  []model.LabelKeyValue{},
			expectErr: false,
		},
		{
			name:      "single pair",
			input:     "key1=value1",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value1"}},
			expectErr: false,
		},
		{
			name:      "multiple pairs",
			input:     "key1=value1;key2=value2;key3=value3",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}, {Key: "key3", Value: "value3"}},
			expectErr: false,
		},
		{
			name:      "same key multiple values",
			input:     "environment=production;environment=staging;region=us-east",
			expected:  []model.LabelKeyValue{{Key: "environment", Value: "production"}, {Key: "environment", Value: "staging"}, {Key: "region", Value: "us-east"}},
			expectErr: false,
		},
		{
			name:      "whitespace handling",
			input:     " key1 = value1 ; key2 = value2 ",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "empty pairs skipped",
			input:     "key1=value1;;key2=value2",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "leading and trailing separators",
			input:     ";key1=value1;key2=value2;",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "empty value",
			input:     "key1=;key2=value2",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: ""}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "value with equals sign",
			input:     "key1=value=with=equals;key2=normal",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value=with=equals"}, {Key: "key2", Value: "normal"}},
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
			expected:  []model.LabelKeyValue{{Key: "", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "whitespace only pairs skipped",
			input:     "key1=value1;   ;key2=value2",
			expected:  []model.LabelKeyValue{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expectErr: false,
		},
		{
			name:      "special characters in key and value",
			input:     "env-name=prod-env;region=us-east-1",
			expected:  []model.LabelKeyValue{{Key: "env-name", Value: "prod-env"}, {Key: "region", Value: "us-east-1"}},
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
