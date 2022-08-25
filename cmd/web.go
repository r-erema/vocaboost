package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jomei/notionapi"
	"github.com/r-erema/vocaboost/internal/application/repository"
	"github.com/r-erema/vocaboost/internal/application/service/dictionary"
	"github.com/r-erema/vocaboost/internal/application/service/images"
	"github.com/r-erema/vocaboost/internal/application/service/spacedrepetition"
	"github.com/r-erema/vocaboost/internal/application/service/textparser"
	"github.com/r-erema/vocaboost/internal/port"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

const (
	envVarRedisHost     = "REDIS_HOST"
	envVarRedisUsername = "REDIS_USERNAME"
	envVarRedisPassword = "REDIS_PASSWORD"

	envVarWordsKey = "WORDS_API_KEY"

	envVarGoogleSearchKey      = "GOOGLE_SEARCH_API_KEY"
	envVarGoogleSearchEngineID = "GOOGLE_SEARCH_ENGINE_ID"

	envVarNotionKey        = "NOTION_API_KEY"
	envVarNotionDatabaseID = "NOTION_DATABASE_ID"

	redisWordsDB = 0
)

type config struct {
	redisHost,
	redisUsername,
	redisPassword,

	wordsAPIKey,

	googleSearchAPIKey,
	googleSearchEngineID,

	notionAPIKey,
	notionDatabaseID string
}

func main() {
	cfg := configFromENVs()

	web := gin.Default()
	web.LoadHTMLGlob("./html_template/*")
	web.StaticFile("/favicon.ico", "./assets/favicon.ico")

	if err := web.SetTrustedProxies(nil); err != nil {
		log.Panicf("setting trusted proxies error: %s", err)
	}

	googleCustomSearchService, err := customsearch.NewService(context.Background(), option.WithAPIKey(cfg.googleSearchAPIKey))
	if err != nil {
		log.Panicf("custom search service creation error: %s", err)
	}

	httpHandler := port.NewHTTPHandler(
		&textparser.V1{},
		repository.NewRedisWordsRepo(
			redisClient(cfg.redisHost, cfg.redisUsername, cfg.redisPassword, redisWordsDB),
		),
		spacedrepetition.NewNotion(
			notionapi.NewClient(notionapi.Token(cfg.notionAPIKey)), notionapi.DatabaseID(cfg.notionDatabaseID),
		),
		dictionary.NewWordsAPI(cfg.wordsAPIKey),
		images.NewGoogleCustomSearch(googleCustomSearchService, cfg.googleSearchEngineID),
	)

	web.GET(port.IndexHTTPPath, httpHandler.Index)
	web.POST(port.IndexHTTPPath, httpHandler.SplitTextToWords)
	web.POST(port.SaveWordsHTTPPath, httpHandler.SaveWords)
	web.POST(port.UploadSpacedRepetitionHTTPPath, httpHandler.UploadToSpacedRepetitionService)

	if err := web.Run(); err != nil {
		log.Panicf("server runnning error: %s", err)
	}
}

func configFromENVs() config {
	var varExists bool

	cfg := config{
		redisHost:            "",
		redisUsername:        "",
		redisPassword:        "",
		wordsAPIKey:          "",
		googleSearchAPIKey:   "",
		googleSearchEngineID: "",
		notionAPIKey:         "",
		notionDatabaseID:     "",
	}

	if cfg.redisHost, varExists = os.LookupEnv(envVarRedisHost); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarRedisHost)
	}

	if cfg.redisUsername, varExists = os.LookupEnv(envVarRedisUsername); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarRedisUsername)
	}

	if cfg.redisPassword, varExists = os.LookupEnv(envVarRedisPassword); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarRedisPassword)
	}

	if cfg.wordsAPIKey, varExists = os.LookupEnv(envVarWordsKey); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarWordsKey)
	}

	if cfg.googleSearchAPIKey, varExists = os.LookupEnv(envVarGoogleSearchKey); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarGoogleSearchKey)
	}

	if cfg.googleSearchEngineID, varExists = os.LookupEnv(envVarGoogleSearchEngineID); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarGoogleSearchEngineID)
	}

	if cfg.notionAPIKey, varExists = os.LookupEnv(envVarNotionKey); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarNotionKey)
	}

	if cfg.notionDatabaseID, varExists = os.LookupEnv(envVarNotionDatabaseID); !varExists {
		log.Panicf("reqiured env var `%s` doesn't exist", envVarNotionDatabaseID)
	}

	return cfg
}

func redisClient(host, username, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{ //nolint: exhaustruct
		Addr:     host,
		Username: username,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Panicf("redis ping error: %s", err)
	}

	return rdb
}
