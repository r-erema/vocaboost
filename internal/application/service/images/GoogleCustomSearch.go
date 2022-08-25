package images

import (
	"context"
	"fmt"

	"google.golang.org/api/customsearch/v1"
)

type GoogleCustomSearch struct {
	service  *customsearch.Service
	engineID string
}

func NewGoogleCustomSearch(service *customsearch.Service, engineID string) *GoogleCustomSearch {
	return &GoogleCustomSearch{service: service, engineID: engineID}
}

func (gcs GoogleCustomSearch) Search(ctx context.Context, words []string) ([]*WordImagesDTO, error) {
	wordImages := make([]*WordImagesDTO, len(words))

	for i := range words {
		resp, err := gcs.service.Cse.List().Context(ctx).SearchType("image").Cx(gcs.engineID).Q(words[i]).Do()
		if err != nil {
			return nil, fmt.Errorf("custom search call error for word `%s`: %w", words[i], err)
		}

		var filteredLinks []string

		for _, item := range resp.Items {
			if item.Mime == "image/png" || item.Mime == "image/jpeg" {
				filteredLinks = append(filteredLinks, item.Link)
			}
		}

		wordImages[i] = NewWordImagesDTO(words[i], filteredLinks)
	}

	return wordImages, nil
}
