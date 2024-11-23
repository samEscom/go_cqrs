package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"sam.com/go/cqrs/models"
)

type ElastiSearchRepository struct {
	client *elastic.Client
}

func NewElastic(url string) (*ElastiSearchRepository, error) {
	client, err := elastic.NewClient(
		elastic.Config{
			Addresses: []string{url},
		})

	if err != nil {
		return nil, err
	}

	return &ElastiSearchRepository{client: client}, nil
}

func (el *ElastiSearchRepository) Close() {
	//
}

func (el *ElastiSearchRepository) IndexFeed(ctx context.Context, feed models.Feed) error {
	body, _ := json.Marshal(feed)

	_, err := el.client.Index(
		"feeds",
		bytes.NewReader(body),
		el.client.Index.WithDocumentID(feed.Id),
		el.client.Index.WithContext(ctx),
		el.client.Index.WithRefresh("wait_for"),
	)

	return err
}

func (el *ElastiSearchRepository) SearchFeed(ctx context.Context, query string) (results []models.Feed, err error) {
	var buf bytes.Buffer

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":            query,
				"fields":           []string{"title", "description"},
				"fuzziness":        3,
				"cutoff_frecuency": 0.0001,
			},
		},
	}

	if err = json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := el.client.Search(
		el.client.Search.WithContext(ctx),
		el.client.Search.WithIndex("feeds"),
		el.client.Search.WithBody(&buf),
		el.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			results = nil
		}
	}()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var eRes map[string]interface{}

	if err := json.NewDecoder(res.Body).Decode(&eRes); err != nil {
		return nil, err
	}

	var feeds []models.Feed

	for _, hit := range eRes["hits"].(map[string]interface{})["hits"].([]interface{}) {
		feed := models.Feed{}

		source := hit.(map[string]interface{})["_source"]

		marshal, err := json.Marshal(source)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(marshal, &feed); err != nil {
			feeds = append(feeds, feed)
		}
	}

	return feeds, nil
}
