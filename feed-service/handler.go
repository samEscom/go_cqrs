package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/ksuid"
	"sam.com/go/cqrs/events"
	"sam.com/go/cqrs/models"
	"sam.com/go/cqrs/repository"
)

type createdFeedRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func createdFeedHandler(w http.ResponseWriter, r *http.Request) {
	var req createdFeedRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdAt := time.Now().UTC()

	id, err := ksuid.NewRandom()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := models.Feed{
		Id:          id.String(),
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   createdAt,
	}

	if err := repository.InsertFeed(r.Context(), &feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := events.PublishCreatedFeed(r.Context(), &feed); err != nil {
		log.Printf("failed to publish created feed event: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Conent-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}
