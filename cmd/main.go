package main

import (
	"embed"
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-app-chaosengine/config"
	"github.com/mattermost/mattermost-app-chaosengine/gameday"
	"github.com/mattermost/mattermost-app-chaosengine/mattermost"
	"github.com/mattermost/mattermost-app-chaosengine/store"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/oklog/oklog/pkg/group"
	log "github.com/sirupsen/logrus"
)

//go:embed manifest.json
var manifestSource []byte //nolint: gochecknoglobals

//go:embed static
var staticAssets embed.FS //nolint: gochecknoglobals

var logger *log.Logger

func main() {
	logger = log.New()
	logger.Out = os.Stdout
	logger.Formatter = &log.JSONFormatter{}
	// Load config
	cfg, err := config.Load(logger)
	if err != nil {
		log.WithError(err).Error("failed to load config")
		os.Exit(1)
		return
	}

	if cfg.Debug {
		logger.SetLevel(log.DebugLevel)
	}

	// apps manifest
	var manifest apps.Manifest
	err = json.Unmarshal(manifestSource, &manifest)
	if err != nil {
		log.WithError(err).Error("failed to load manfest")
		os.Exit(1)
		return
	}

	store, err := store.New(cfg.Database, logger)
	if err != nil {
		log.WithError(err).Error("failed to connect to Database")
		os.Exit(1)
		return
	}
	// Run migrations on startup
	err = store.Migrate()
	if err != nil {
		log.WithError(err).Error("failed to run migrations")
		os.Exit(1)
		return
	}
	// Create HTTP router
	r := mux.NewRouter()
	r.Use(logRequest)

	// mattermost App settings + Routes
	manifest.HTTPRootURL = cfg.App.RootURL
	manifest.AppType = cfg.App.Type
	mattermost.AddRoutes(r, &manifest, staticAssets, cfg.App.Secret, cfg.Debug)

	// Gameday Routes
	gamedayRepo := gameday.NewRepository(store)
	gamedaySvc := gameday.NewService(gamedayRepo)
	gameday.AddRoutes(r, gamedaySvc, logger)

	httpListener, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		logger.WithError(err).Errorf("failed to listen in %s", cfg.ListenAddress)
		os.Exit(1)
	}

	var g group.Group
	g.Add(func() error {
		logger.WithField("listen", cfg.ListenAddress).Info("Server started")
		return http.Serve(httpListener, r)
	}, func(error) {
		httpListener.Close()
	})

	logger.WithError(g.Run()).Error("exit")
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithField("method", r.Method).
			WithField("url", r.URL.Path).
			Info("received HTTP request")
		next.ServeHTTP(w, r)
	})
}
