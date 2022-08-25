package port

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/r-erema/vocaboost/internal/application/repository"
	"github.com/r-erema/vocaboost/internal/application/service/dictionary"
	"github.com/r-erema/vocaboost/internal/application/service/images"
	"github.com/r-erema/vocaboost/internal/application/service/spacedrepetition"
	"github.com/r-erema/vocaboost/internal/application/service/textparser"
	"github.com/r-erema/vocaboost/internal/domain"
	"github.com/thoas/go-funk"
)

const (
	userErrSomethingWentWrong  = "something went wrong"
	saveTargetRepoKnownWords   = "known_words"
	saveTargetRepoIgnoredWords = "ignored_words"

	IndexHTTPPath                  = "/"
	SaveWordsHTTPPath              = "/save-words"
	UploadSpacedRepetitionHTTPPath = "/upload-spaced-repetition"

	maxDefinitions = 4
	maxImages      = 8
	maxExamples    = 4
)

var (
	errBadImageAndDefinitionWord = errors.New("images word and definitions word aren't equal")
	errBadImagesCount            = errors.New("bad images count")
)

type HTTPHandler struct {
	textParser              textparser.Interface
	wordsRepo               repository.Interface
	spacedRepetitionService spacedrepetition.Interface
	dictionary              dictionary.Interface
	images                  images.Interface
}

func NewHTTPHandler(
	textParser textparser.Interface,
	wordsTracker repository.Interface,
	spacedRepetitionService spacedrepetition.Interface,
	dictionaryService dictionary.Interface,
	imagesService images.Interface,
) *HTTPHandler {
	return &HTTPHandler{
		textParser:              textParser,
		wordsRepo:               wordsTracker,
		spacedRepetitionService: spacedRepetitionService,
		dictionary:              dictionaryService,
		images:                  imagesService,
	}
}

func (hh *HTTPHandler) Index(context *gin.Context) {
	knownWordsCount, err := hh.wordsRepo.KnownWordsCount(context.Request.Context())
	if err != nil {
		log.Printf("getting known words count error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	ignoredWordsCount, err := hh.wordsRepo.IgnoredWordsCount(context.Request.Context())
	if err != nil {
		log.Printf("getting ignored words count error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	context.HTML(http.StatusOK, "index.html", gin.H{
		"known_words_count":   knownWordsCount,
		"ignored_words_count": ignoredWordsCount,
	})
}

func (hh *HTTPHandler) SplitTextToWords(context *gin.Context) {
	form := new(struct {
		Text string `form:"text"`
	})

	if err := context.ShouldBind(form); err != nil {
		log.Printf("form binding error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	words, err := hh.textParser.ExtractWords(form.Text)
	if err != nil {
		log.Printf("extrating words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	words, ok := funk.Map(words, strings.ToLower).([]string)
	if !ok {
		log.Printf("unable to assert to string slice")
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	words = funk.FilterString(words, func(w string) bool {
		if _, err = strconv.Atoi(w); err == nil {
			return false
		}

		return len(w) > 1
	})

	words = funk.UniqString(words)

	if words, err = hh.wordsRepo.FilterKnownWords(context.Request.Context(), words); err != nil {
		log.Printf("filtering known words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	if words, err = hh.wordsRepo.FilterIgnoredWords(context.Request.Context(), words); err != nil {
		log.Printf("filtering ignored words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	context.HTML(http.StatusOK, "words_list.html", gin.H{
		"save_words_http_path": SaveWordsHTTPPath,
		"index_http_path":      IndexHTTPPath,

		"words": words,

		"known_words_value":   saveTargetRepoKnownWords,
		"ignored_words_value": saveTargetRepoIgnoredWords,
	})
}

func (hh *HTTPHandler) SaveWords(context *gin.Context) {
	err := context.Request.ParseForm()
	if err != nil {
		log.Printf("parse form error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	var knownWordsToSave, ignoredWordsToSave, unknownWords []string

	for word, saveTargetValues := range context.Request.PostForm {
		switch saveTargetValues[0] {
		case saveTargetRepoKnownWords:
			knownWordsToSave = append(knownWordsToSave, word)
		case saveTargetRepoIgnoredWords:
			ignoredWordsToSave = append(ignoredWordsToSave, word)
		default:
			unknownWords = append(unknownWords, word)
		}
	}

	if err = hh.wordsRepo.SaveAsKnown(context, knownWordsToSave); err != nil {
		log.Printf("saving known words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	if err = hh.wordsRepo.SaveAsIgnored(context, ignoredWordsToSave); err != nil {
		log.Printf("saving ignored words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	context.HTML(http.StatusOK, "unknown_words_list.html", gin.H{
		"index_http_path":          IndexHTTPPath,
		"upload_spaced_repetition": UploadSpacedRepetitionHTTPPath,

		"unknown_words": unknownWords,
	})
}

func (hh *HTTPHandler) UploadToSpacedRepetitionService(context *gin.Context) {
	form := struct {
		UnknownWordsText string `form:"unknown_words"`
	}{}

	if err := context.Bind(&form); err != nil {
		log.Printf("parse form error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	unknownWords, err := hh.textParser.ExtractWords(form.UnknownWordsText)
	if err != nil {
		log.Printf("extracting words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	wordsDefinitions, err := hh.dictionary.WordsInfo(context.Request.Context(), unknownWords)
	if err != nil {
		log.Printf("getting words info error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	wordsImages, err := hh.images.Search(context.Request.Context(), unknownWords)
	if err != nil {
		log.Printf("getting words images error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	wordsImagesClarified, err := hh.images.Search(context.Request.Context(), funk.Map(unknownWords, func(word string) string {
		return word + " meaning"
	}).([]string))
	if err != nil {
		log.Printf("getting clarified words images error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	words, err := prepareWords(unknownWords, wordsDefinitions, wordsImages, wordsImagesClarified)
	if err != nil {
		log.Printf("preparing words error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	if err = hh.spacedRepetitionService.UploadWords(context.Request.Context(), words); err != nil {
		log.Printf("uploading plainWords to the spaced repetiotion service error: %s", err)
		context.String(http.StatusInternalServerError, userErrSomethingWentWrong)

		return
	}

	context.HTML(http.StatusOK, "result.html", gin.H{
		"index_http_path": IndexHTTPPath,
	})
}

func prepareWords(
	unknownWords []string,
	wordsDefinitions []*dictionary.WordInfoDTO,
	wordsImages,
	wordsImagesClarified []*images.WordImagesDTO,
) ([]*domain.Word, error) {
	words := make([]*domain.Word, len(unknownWords))

	for i := 0; i < len(unknownWords); i++ {
		definition := wordsDefinitions[i]
		wordImages := wordsImages[i]
		wordImagesClarified := wordsImagesClarified[i]

		definitionAndImagesCompatible := definition.Word() == wordImages.Word()
		if !definitionAndImagesCompatible {
			return nil, fmt.Errorf(
				"%w, images word: %s, definitions word: %s",
				errBadImageAndDefinitionWord,
				wordImages.Word(),
				definition.Word(),
			)
		}

		definitions := definition.Definitions()
		if len(definitions) > maxDefinitions {
			definitions = definitions[:maxDefinitions]
		}

		examples := definition.Examples()
		if len(examples) > maxExamples {
			examples = examples[:maxExamples]
		}

		imageURLs, err := prepareImages(wordImages.Urls(), wordImagesClarified.Urls())
		if err != nil {
			return nil, fmt.Errorf("preparing images for word `%s` error: %w", definition.Word(), err)
		}

		words[i] = domain.NewWord(definition.Word(), definitions, examples, imageURLs)
	}

	return words, nil
}

func prepareImages(normalImages, clarifiedImages []string) ([]string, error) {
	requiredImagesCountPerType := maxImages / 2
	if len(normalImages) < requiredImagesCountPerType || len(clarifiedImages) < requiredImagesCountPerType {
		return nil, errBadImagesCount
	}

	result := append(normalImages[:requiredImagesCountPerType], clarifiedImages[:requiredImagesCountPerType]...)

	return result, nil
}
