package aws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getAddr(t *testing.T) {
	tcs := []struct {
		host     string
		expected string
	}{
		{"https://test.com", "test.com:443"},
		{"http://test.com", "test.com:80"},
		{"test.com", "test.com:443"},
		{"https://test.com/some/path", "test.com:443"},
		{"http://test.com/some/path", "test.com:80"},
		{"test.com/some/path", "test.com:443"},
	}

	for _, tc := range tcs {
		actual, err := getHostAndPort(tc.host)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, actual, fmt.Sprintf("host: %s", tc.host))
	}
}
