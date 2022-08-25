package images

import "context"

type WordImagesDTO struct {
	word string
	urls []string
}

func (w WordImagesDTO) Word() string {
	return w.word
}

func (w WordImagesDTO) Urls() []string {
	return w.urls
}

func NewWordImagesDTO(word string, urls []string) *WordImagesDTO {
	return &WordImagesDTO{word: word, urls: urls}
}

type Interface interface {
	Search(ctx context.Context, words []string) ([]*WordImagesDTO, error)
}
