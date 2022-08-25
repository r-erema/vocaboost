package spacedrepetition

import (
	"context"

	"github.com/r-erema/vocaboost/internal/domain"
)

type Interface interface {
	UploadWords(ctx context.Context, words []*domain.Word) error
}
