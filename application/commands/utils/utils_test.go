package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSliceFlag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", nil},
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
		{"empty string", "", nil, false},
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
