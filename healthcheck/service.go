package healthcheck

import (
	"net/http"
)

type Service struct {
	healthChecks []func() error
}

func New(healthChecks ...func() error) *Service {
	return &Service{
		healthChecks: healthChecks,
	}
}

func (s *Service) HealthHandler(w http.ResponseWriter, req *http.Request) {
	writtenHeader := false

	for _, check := range s.healthChecks {
		if err := check(); err != nil {
			if !writtenHeader {
				w.WriteHeader(http.StatusInternalServerError)
				writtenHeader = true
			}
			w.Write([]byte(err.Error())) // nolint
			w.Write([]byte("\n\n"))      // nolint
		}
	}

	if !writtenHeader {
		w.WriteHeader(http.StatusOK)
	}
}
