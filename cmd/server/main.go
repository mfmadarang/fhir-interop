package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/mfmadarang/fhir-interop/internal/auth"
	"github.com/mfmadarang/fhir-interop/internal/demo"
	"github.com/mfmadarang/fhir-interop/internal/graph"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	db, err := store.Connect()
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	if err := store.Migrate(db); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	resolver := &graph.Resolver{DB: db}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	demoHandler := demo.NewHandler(db, terminology.NewClient())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/", playground.Handler("fhir-interop GraphQL playground", "/query"))
	http.Handle("/query", auth.APIKeyMiddleware(apiKey)(srv))
	http.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("cmd/server/web/index.html")
		if err != nil {
			http.Error(w, "app not found: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})
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

	log.Printf("listening on :%s (playground at http://localhost:%s/, browser at http://localhost:%s/app, demo at http://localhost:%s/demo)", port, port, port, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
