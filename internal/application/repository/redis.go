package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/thoas/go-funk"
)

const (
	knownWordsPrefix   = "k:"
	ignoredWordsPrefix = "i:"
)

type RedisWordsRepo struct {
	client *redis.Client
}

func NewRedisWordsRepo(clientKnownWords *redis.Client) *RedisWordsRepo {
	return &RedisWordsRepo{client: clientKnownWords}
}

func (rwr RedisWordsRepo) SaveAsKnown(ctx context.Context, words []string) error {
	if len(words) == 0 {
		return nil
	}

	values := make(map[string]string, len(words))
	for _, word := range words {
		values[knownWordKey(word)] = word
	}

	if err := rwr.client.MSet(ctx, values).Err(); err != nil {
		return fmt.Errorf("redis MSet operation error: %w", err)
	}

	return nil
}

func (rwr RedisWordsRepo) SaveAsIgnored(ctx context.Context, words []string) error {
	if len(words) == 0 {
		return nil
	}

	values := make(map[string]string, len(words))
	for _, word := range words {
		values[ignoredWordKey(word)] = word
	}

	if err := rwr.client.MSet(ctx, values).Err(); err != nil {
		return fmt.Errorf("redis MSet operation error: %w", err)
	}

	return nil
}

var errToStringAssertion = errors.New("unable assert to strings slice")

func (rwr RedisWordsRepo) FilterKnownWords(ctx context.Context, words []string) ([]string, error) {
	if keys, ok := funk.Map(words, knownWordKey).([]string); ok {
		return rwr.filterWords(ctx, words, keys)
	}

	return nil, errToStringAssertion
}

func (rwr RedisWordsRepo) FilterIgnoredWords(ctx context.Context, words []string) ([]string, error) {
	if keys, ok := funk.Map(words, ignoredWordKey).([]string); ok {
		return rwr.filterWords(ctx, words, keys)
	}

	return nil, errToStringAssertion
}

func (rwr RedisWordsRepo) filterWords(ctx context.Context, words, keys []string) ([]string, error) {
	var filteredWords []string

	for i := range keys {
		count, err := rwr.client.Exists(ctx, keys[i]).Result()
		if err != nil {
			return nil, fmt.Errorf("`Exists` operation error: %w", err)
		}

		if count < 1 {
			filteredWords = append(filteredWords, words[i])
		}
	}

	return filteredWords, nil
}

func (rwr RedisWordsRepo) KnownWordsCount(ctx context.Context) (int, error) {
	keys, err := rwr.client.Keys(ctx, fmt.Sprintf("%s*", knownWordsPrefix)).Result()
	if err != nil {
		return -1, fmt.Errorf("getting keys error: %w", err)
	}

	return len(keys), nil
}

func (rwr RedisWordsRepo) IgnoredWordsCount(ctx context.Context) (int, error) {
	keys, err := rwr.client.Keys(ctx, fmt.Sprintf("%s*", ignoredWordsPrefix)).Result()
	if err != nil {
		return -1, fmt.Errorf("getting keys error: %w", err)
	}

	return len(keys), nil
}

func knownWordKey(key string) string {
	return fmt.Sprintf("%s%s", knownWordsPrefix, key)
}

func ignoredWordKey(key string) string {
	return fmt.Sprintf("%s%s", ignoredWordsPrefix, key)
}
