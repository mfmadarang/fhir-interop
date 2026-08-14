package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/mfmadarang/fhir-interop/internal/auth"
	"github.com/mfmadarang/fhir-interop/internal/graph"
	"github.com/mfmadarang/fhir-interop/internal/store"
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

	log.Printf("listening on :%s (playground at http://localhost:%s/, browser at http://localhost:%s/app)", port, port, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
