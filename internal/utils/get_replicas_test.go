package utils_test

import (
	"testing"
	"wildwest/internal/utils"

	"github.com/stretchr/testify/assert"
)

const ReplicasEnvKey = "REPLICAS"

func TestGetReplicas(t *testing.T) {
	tests := []struct {
		name        string
		envKeyValue string
		want        int
		err         error
	}{
		{"10 replicas", "10", 10, nil},
		{"2 replicas", "2", 2, nil},
		{"1 replica", "1", 0, utils.ErrNotEnoughReplicas},
		{"0 replicas", "0", 0, utils.ErrNotEnoughReplicas},
		{"incorrect value", "foo", 0, utils.ErrIncorrectEnvVariable},
		{"empty environment variable value", "", 0, utils.ErrIncorrectEnvVariable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(ReplicasEnvKey, tc.envKeyValue)

			got, err := utils.GetReplicas(ReplicasEnvKey)
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetReplicasNoEnvVariable(t *testing.T) {
	got, err := utils.GetReplicas(ReplicasEnvKey)
	assert.ErrorIs(t, err, utils.ErrEnvVariableNotFound)
	assert.Equal(t, 0, got)
}
