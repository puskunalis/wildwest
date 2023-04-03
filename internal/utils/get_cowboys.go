package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrCowboyListTooShort      = errors.New("cowboy list size is smaller than cowboy replica count")
	ErrCowboyNamesNotUnique    = errors.New("cowboy names are not unique")
	ErrCowboyHealthNotPositive = errors.New("cowboy health must be positive")
	ErrCowboyDamageNotPositive = errors.New("cowboy damage must be positive")
)

type cowboyListValidationFunc func([]Cowboy) error

// GetCowboys returns a validated list of cowboys from a given file and replica count
func GetCowboys(filename string, replicas int) ([]Cowboy, error) {
	cowboys, err := loadCowboys(filename)
	if err != nil {
		return nil, fmt.Errorf("get cowboys: %w", err)
	}

	// validate the cowboy list
	if err := validateCowboyList(cowboys, replicas); err != nil {
		return nil, fmt.Errorf("get cowboys: %w", err)
	}

	return cowboys, nil
}

// loadCowboys loads a list of cowboys from a given filename
func loadCowboys(filename string) ([]Cowboy, error) {
	// open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// read file
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var cowboys []Cowboy

	// unmarshal JSON
	err = json.Unmarshal(bytes, &cowboys)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}

	return cowboys, nil
}

// buildCompareCowboyListToReplicaCount builds a function to check cowboy list length given replica count
func buildCompareCowboyListToReplicaCount(replicas int) cowboyListValidationFunc {
	return func(cowboys []Cowboy) error {
		if len(cowboys) < replicas {
			return ErrCowboyListTooShort
		}

		return nil
	}
}

// validateCowboyList validates the cowboy list values
func validateCowboyList(cowboys []Cowboy, replicas int) error {
	compareCowboyListToReplicaCount := buildCompareCowboyListToReplicaCount(replicas)

	validationFuncs := []cowboyListValidationFunc{
		compareCowboyListToReplicaCount,
		areCowboyNamesUnique,
		areCowboyHealthValuesPositive,
		areCowboyDamageValuesPositive,
	}

	for _, f := range validationFuncs {
		if err := f(cowboys); err != nil {
			return fmt.Errorf("validate cowboy list: %w", err)
		}
	}

	return nil
}

// areCowboyNamesUnique checks whether all cowboys have unique names
func areCowboyNamesUnique(cowboys []Cowboy) error {
	nameSet := make(map[string]struct{})
	for _, cowboy := range cowboys {
		if _, ok := nameSet[cowboy.Name]; ok {
			return ErrCowboyNamesNotUnique
		}

		nameSet[cowboy.Name] = struct{}{}
	}

	return nil
}

// areCowboyHealthValuesPositive checks whether cowboys have positive health values
func areCowboyHealthValuesPositive(cowboys []Cowboy) error {
	for _, cowboy := range cowboys {
		if cowboy.Health <= 0 {
			return ErrCowboyHealthNotPositive
		}
	}

	return nil
}

// areCowboyDamageValuesPositive checks whether all cowboys have positive damage values
func areCowboyDamageValuesPositive(cowboys []Cowboy) error {
	for _, cowboy := range cowboys {
		if cowboy.Damage <= 0 {
			return ErrCowboyDamageNotPositive
		}
	}

	return nil
}
