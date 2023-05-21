package utils

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// StartReadinessServer starts a server on /ready, call is blocking
func StartReadinessServer(logger *zap.Logger, port int) {
	http.HandleFunc("/ready", readyHandler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		logger.Fatal("readiness server listen", zap.Error(err))
	}
}
