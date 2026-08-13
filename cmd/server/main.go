package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/mfmadarang/fhir-interop/internal/graph"
	"github.com/mfmadarang/fhir-interop/internal/store"
)

func main() {
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
	http.Handle("/query", srv)

	log.Printf("listening on :%s (playground at http://localhost:%s/)", port, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
