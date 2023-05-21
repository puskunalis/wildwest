package utils_test

import (
	"fmt"
	"testing"
	"wildwest/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestGetID(t *testing.T) {
	tests := []struct {
		hostname string
		want     int
		err      error
	}{
		{"cowboy-0", 0, nil},
		{"cowboy-1", 1, nil},
		{"cowboy-2", 2, nil},
		{"cowboy-10", 10, nil},
		{"cowboy-100", 100, nil},
		{"c-1", 1, nil},
		{"c-10", 10, nil},
		{"-1", 0, utils.ErrInvalidHostnameFormat},
		{"c-", 0, utils.ErrInvalidHostnameFormat},
		{"c--", 0, utils.ErrInvalidHostnameFormat},
		{"-", 0, utils.ErrInvalidHostnameFormat},
		{"", 0, utils.ErrInvalidHostnameFormat},
		{"cowboy", 0, utils.ErrInvalidHostnameFormat},
		{"cowboy-", 0, utils.ErrInvalidHostnameFormat},
		{"cowboy-1-1", 0, utils.ErrInvalidHostnameFormat},
		{"cowboy-cowboy", 0, utils.ErrInvalidHostnameFormat},
		{"cowboy-c", 0, utils.ErrInvalidHostnameFormat},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("hostname '%s'", tc.hostname), func(t *testing.T) {
			getHostnameFunc := func() (string, error) {
				return tc.hostname, nil
			}

			got, err := utils.GetID(getHostnameFunc)
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}
