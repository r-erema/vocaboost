package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	definitionsAPICallURLTemplate = "https://%s/words/%s/definitions"
	examplesAPICallURLTemplate    = "https://%s/words/%s/examples"
	apiHost                       = "wordsapiv1.p.rapidapi.com"
)

type WordsAPI struct {
	apiKey string
}

type definitionsResponse struct {
	Word        string `json:"word"`
	Definitions []struct {
		Definition   string `json:"definition"`
		PartOfSpeech string `json:"partOfSpeech"` //nolint:tagliatelle
	} `json:"definitions"`
}

type examplesResponse struct {
	Word     string   `json:"word"`
	Examples []string `json:"examples"`
}

func NewWordsAPI(apiKey string) *WordsAPI {
	return &WordsAPI{apiKey: apiKey}
}

func (wa WordsAPI) WordsInfo(ctx context.Context, words []string) ([]*WordInfoDTO, error) {
	wordsInfo := make([]*WordInfoDTO, len(words))

	for i := range words {
		definitionsResp, err := wa.definitionsAPICall(ctx, words[i])
		if err != nil {
			return nil, fmt.Errorf("definition API call error: %w", err)
		}

		definitions := make([]*DefinitionDTO, len(definitionsResp.Definitions))
		for j := range definitionsResp.Definitions {
			definitions[j] = NewDefinition(definitionsResp.Definitions[j].Definition, definitionsResp.Definitions[j].PartOfSpeech)
		}

		examples, err := wa.examplesAPICall(ctx, words[i])
		if err != nil {
			return nil, fmt.Errorf("example API call error: %w", err)
		}

		wordsInfo[i] = NewWordInfoDTO(words[i], definitions, examples.Examples)
	}

	return wordsInfo, nil
}

func (wa WordsAPI) definitionsAPICall(ctx context.Context, word string) (*definitionsResponse, error) {
	url := fmt.Sprintf(definitionsAPICallURLTemplate, apiHost, word)
	definitionItems := new(definitionsResponse)

	if err := wa.apiCall(ctx, url, definitionItems); err != nil {
		return nil, fmt.Errorf("API call(word `%s`) error: %w", word, err)
	}

	return definitionItems, nil
}

func (wa WordsAPI) examplesAPICall(ctx context.Context, word string) (*examplesResponse, error) {
	url := fmt.Sprintf(examplesAPICallURLTemplate, apiHost, word)
	exampleItems := new(examplesResponse)

	if err := wa.apiCall(ctx, url, exampleItems); err != nil {
		return nil, fmt.Errorf("API call(word `%s`) error: %w", word, err)
	}

	return exampleItems, nil
}

func (wa WordsAPI) apiCall(ctx context.Context, url string, dto interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("creation request error: %w", err)
	}

	req.Header.Add("X-RapidAPI-Key", wa.apiKey)
	req.Header.Add("X-RapidAPI-Host", apiHost)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request execution error: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("body closing error: %s", err.Error())
		}
	}()

	if err = json.NewDecoder(resp.Body).Decode(dto); err != nil {
		return fmt.Errorf("response decoding to examples error: %w", err)
	}

	return nil
}
