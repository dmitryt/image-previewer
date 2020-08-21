package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitryt/image-previewer/internal/config"
	"github.com/dmitryt/image-previewer/internal/fetcher"
	"github.com/dmitryt/image-previewer/internal/resizer"
	"github.com/dmitryt/image-previewer/internal/transport"
	"github.com/dmitryt/image-previewer/internal/utils"
	"github.com/rs/zerolog/log"
)

var (
	ErrImageFetch         = errors.New("error during fetching the image")
	ErrImageResize        = errors.New("error during resizing the image")
	ErrInvalidURI         = errors.New("invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
	ErrImageCopyFromCache = errors.New("error during copying the image from cache")
)

type App struct {
	config    *config.Config
	resizer   *resizer.Resizer
	transport *transport.Transport
}

type DummyResponse struct {
	OK bool
}

func New(config *config.Config, client *http.Client) (*App, error) {
	rsz, err := resizer.New(config)
	transport := transport.New(fetcher.NewHTTPFetcher(client, config), rsz)

	return &App{
		config:    config,
		resizer:   rsz,
		transport: transport,
	}, err
}

func (p *App) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	dummyResponse := DummyResponse{true}

	content, err := json.Marshal(dummyResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(content)
}

func (p *App) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	urlParams := utils.ParseURL("/" + r.URL.Path[1:])
	log.Debug().Msgf("url params %+v", urlParams)
	if urlParams.Error != nil {
		log.Error().Msgf("%s: %+v", ErrInvalidURI, urlParams)
		w.WriteHeader(400)
		fmt.Fprintf(w, "%s", ErrInvalidURI)

		return
	}
	if !p.resizer.HasFile(urlParams) {
		log.Debug().Msg("File was not found in cache, fetching the content...")
		statusCode, content, err := p.transport.Receive(urlParams, r.Header)
		if err != nil {
			log.Error().Msgf("%s: %s", ErrImageFetch, err)
			w.WriteHeader(statusCode)
			fmt.Fprint(w, content)

			return
		}
	}

	err := p.transport.Send(urlParams, w)
	if err != nil {
		log.Error().Msgf("%s: %s", ErrImageCopyFromCache, err)
		fmt.Fprintf(w, "%s", ErrImageCopyFromCache)
	}
}

func (p *App) Run(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health-check", p.HealthCheckHandler)
	mux.HandleFunc("/fill/", p.ResizeHandler)

	log.Info().Msgf("Listening at %s", addr)

	return http.ListenAndServe(addr, mux)
}
