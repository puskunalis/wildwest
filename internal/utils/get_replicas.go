package utils

import (
	"os"
	"strconv"
)

const (
	ErrEnvVariableNotFound  = ConstError("given environment variable not found")
	ErrNotEnoughReplicas    = ConstError("must have at least 2 replicas")
	ErrIncorrectEnvVariable = ConstError("given environment variable has incorrect value")
)

// GetReplicas returns the replica count from a given environment variable
func GetReplicas(envKey string) (int, error) {
	// get replicas count string from an environment variable
	replicasStr, ok := os.LookupEnv(envKey)
	if !ok {
		return 0, ErrEnvVariableNotFound
	}

	// convert replicas count string to int
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil {
		return 0, ErrIncorrectEnvVariable
	}

	// check if we have enough replicas
	if replicas <= 1 {
		return 0, ErrNotEnoughReplicas
	}

	return replicas, nil
}
