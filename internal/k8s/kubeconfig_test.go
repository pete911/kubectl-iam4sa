package k8s

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getFlagValue(t *testing.T) {
	tcs := []struct {
		args     []string
		flag     string
		expected string
	}{
		{[]string{"--profile", "default", "--region", "eu-west-2"}, "--profile", "default"},
		{[]string{"--profile", "default", "--region", "eu-west-2"}, "--region", "eu-west-2"},
		{[]string{"--profile", "default", "--region"}, "--region", ""},
	}

	for _, tc := range tcs {
		actual := getFlagValue(tc.args, tc.flag)
		assert.Equal(t, tc.expected, actual, fmt.Sprintf("args: %v flag %s", tc.args, tc.flag))
	}
}
