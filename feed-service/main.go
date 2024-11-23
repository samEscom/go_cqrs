package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"sam.com/go/cqrs/database"
	"sam.com/go/cqrs/events"
	"sam.com/go/cqrs/repository"
)

type Config struct {
	PostgresDB       string `envconfig:"POSTGRES_DB"`
	PostgresUser     string `envconfig:"POSTGRES_USER"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress      string `envconfig:"NATS_ADDRESS"`
}

func newRouter() (router *mux.Router) {
	router = mux.NewRouter()

	router.HandleFunc("/feeds", createdFeedHandler).Methods(http.MethodGet)

	return
}

func main() {
	var cfg Config

	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatalf("%V", err)
	}

	addr := fmt.Sprintf("postgres://%s:%s@postgres/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)

	repo, err := database.NewPostgresRepository(addr)

	if err != nil {
		log.Fatal(err)
	}

	repository.SetRepository(repo)

	n, err := events.NewNats(fmt.Sprintf("nats://%s", cfg.NatsAddress))

	if err != nil {
		log.Fatal(err)
	}

	events.SetEventStore(n)

	defer events.Close()

	router := newRouter()

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}

}
