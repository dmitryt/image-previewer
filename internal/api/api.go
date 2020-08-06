package api

import (
	"net/http"
)

type API struct {
	Client *http.Client
}

func Handlers() http.Handler {
	r := http.NewServeMux()
	api := API{Client: http.DefaultClient}
	r.HandleFunc("/health-check", api.HealthCheckHandler)
	r.HandleFunc("/fill/", api.ResizeHandler)
	return r
}
