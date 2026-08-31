package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/mfmadarang/fhir-interop/internal/auth"
	"github.com/mfmadarang/fhir-interop/internal/config"
	"github.com/mfmadarang/fhir-interop/internal/demo"
	"github.com/mfmadarang/fhir-interop/internal/graph"
	"github.com/mfmadarang/fhir-interop/internal/obs"
	"github.com/mfmadarang/fhir-interop/internal/rest"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// logger isn't set up yet, so print and bail
		slog.Error("loading config", "err", err)
		os.Exit(1)
	}

	logger := obs.NewLogger(cfg)
	slog.SetDefault(logger)

	db, err := store.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("connecting to database", "err", err)
		os.Exit(1)
	}

	if err := store.Migrate(db); err != nil {
		logger.Error("running migrations", "err", err)
		os.Exit(1)
	}

	resolver := &graph.Resolver{DB: db}
	gqlSrv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	demoHandler := demo.NewHandler(db, terminology.NewClient())

	web, err := webHandler()
	if err != nil {
		logger.Error("loading web assets", "err", err)
		os.Exit(1)
	}

	metrics := obs.NewMetrics()

	restHandler := rest.New(rest.NewGormStore(db))

	demoPage := func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("cmd/server/web/pipeline.html")
		if err != nil {
			http.Error(w, "demo page not found: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	}

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("fhir-interop GraphQL playground", "/query"))
	mux.Handle("/query", metrics.Middleware("/query", auth.APIKeyMiddleware(cfg.APIKey)(gqlSrv)))
	mux.Handle("/fhir/", metrics.Middleware("/fhir/", restHandler.Routes()))
	mux.Handle("/app", http.RedirectHandler("/app/", http.StatusMovedPermanently))
	mux.Handle("/app/", metrics.Middleware("/app/", web))
	mux.Handle("/demo", metrics.Middleware("/demo", http.HandlerFunc(demoPage)))
	mux.Handle("/demo/run", metrics.Middleware("/demo/run", http.HandlerFunc(demoHandler.HandleRun)))
	mux.HandleFunc("/demo/stream", demoHandler.HandleStream)
	mux.HandleFunc("/healthz", obs.Healthz)
	mux.Handle("/readyz", obs.Readyz(db))
	mux.Handle("/metrics", metrics.Handler())

	quiet := []string{"/healthz", "/readyz", "/metrics"}
	root := obs.RequestLogger(logger, quiet, mux)

	addr := ":" + cfg.Port
	logger.Info("starting server",
		"addr", addr,
		"playground", "http://localhost:"+cfg.Port+"/",
		"browser", "http://localhost:"+cfg.Port+"/app/",
		"demo", "http://localhost:"+cfg.Port+"/demo",
	)
	if err := http.ListenAndServe(addr, root); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
