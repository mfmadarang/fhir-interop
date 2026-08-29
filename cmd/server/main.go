package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/mfmadarang/fhir-interop/internal/auth"
	"github.com/mfmadarang/fhir-interop/internal/config"
	"github.com/mfmadarang/fhir-interop/internal/demo"
	"github.com/mfmadarang/fhir-interop/internal/graph"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := store.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	if err := store.Migrate(db); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	resolver := &graph.Resolver{DB: db}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	demoHandler := demo.NewHandler(db, terminology.NewClient())

	web, err := webHandler()
	if err != nil {
		log.Fatalf("loading web assets: %v", err)
	}

	http.Handle("/", playground.Handler("fhir-interop GraphQL playground", "/query"))
	http.Handle("/query", auth.APIKeyMiddleware(cfg.APIKey)(srv))
	http.Handle("/app", http.RedirectHandler("/app/", http.StatusMovedPermanently))
	http.Handle("/app/", web)
	http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("cmd/server/web/pipeline.html")
		if err != nil {
			http.Error(w, "demo page not found: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})
	http.HandleFunc("/demo/run", demoHandler.HandleRun)
	http.HandleFunc("/demo/stream", demoHandler.HandleStream)

	addr := ":" + cfg.Port
	log.Printf("listening on %s (playground at http://localhost:%s/, browser at http://localhost:%s/app/, demo at http://localhost:%s/demo)", addr, cfg.Port, cfg.Port, cfg.Port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
