package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const ErrInvalidHostnameFormat = ConstError("invalid hostname format")

type GetHostnameFunc func() (hostname string, err error)

// GetID gets our pod ID from the StatefulSet hostname format (pod-0, pod-1, pod-2...)
func GetID(getHostnameFunc GetHostnameFunc) (int, error) {
	// get hostname
	hostname, err := getHostnameFunc()
	if err != nil {
		return 0, fmt.Errorf("get hostname: %w", err)
	}

	os.Hostname()

	// get id from hostname
	split := strings.Split(hostname, "-")
	if len(split) != 2 {
		return 0, ErrInvalidHostnameFormat
	}
	if len(split[0]) == 0 || len(split[1]) == 0 {
		return 0, ErrInvalidHostnameFormat
	}

	// retrieve id
	idStr := split[1]

	// convert id to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, ErrInvalidHostnameFormat
	}

	return id, nil
}
