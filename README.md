# Go Template Service
### .helm
helm configuration
### cache
cache client
### cmd
app binary
### config
app config, flags and env vars
### healthcheck
health check service with HTTP handler. It handles all checks across other packages (e.g. Mongo, Postgres, Redis)
### http
http server
### http_client
http client
### models
domain-level, not storage specific models
### prometheus
Prometheus metrics and middlewares
### service
business logic service
### storage
storage implementations (e.g. Mongo, Postgres)