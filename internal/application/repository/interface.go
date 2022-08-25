package repository

import "context"

type Interface interface {
	FilterKnownWords(ctx context.Context, words []string) ([]string, error)
	FilterIgnoredWords(ctx context.Context, words []string) ([]string, error)
	SaveAsKnown(ctx context.Context, words []string) error
	SaveAsIgnored(ctx context.Context, words []string) error
	KnownWordsCount(ctx context.Context) (int, error)
	IgnoredWordsCount(ctx context.Context) (int, error)
}
