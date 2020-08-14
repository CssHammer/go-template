package prometheus

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Middlewares struct {
	observer prometheus.Observer
	counter  prometheus.Counter
}

func NewMiddlewares(observer prometheus.Observer, counter prometheus.Counter) *Middlewares {
	return &Middlewares{
		observer: observer,
		counter:  counter,
	}
}

func (m *Middlewares) Timer(next http.Handler) http.Handler {
	if m.observer == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		go m.observer.Observe(time.Since(start).Seconds())
	})
}

func (m *Middlewares) Counter(next http.Handler) http.Handler {
	if m.counter == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go m.counter.Inc()
		next.ServeHTTP(w, r)
	})
}
