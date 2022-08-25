package dictionary

import "context"

type (
	WordInfoDTO struct {
		word           string
		definitionsDTO []*DefinitionDTO
		examples       []string
	}

	DefinitionDTO struct {
		definition   string
		partOfSpeech string
	}
)

func (w WordInfoDTO) Word() string {
	return w.word
}

func (w WordInfoDTO) Definitions() []string {
	slice := make([]string, len(w.definitionsDTO))
	for i := range w.definitionsDTO {
		slice[i] = w.definitionsDTO[i].definition
	}

	return slice
}

func (w WordInfoDTO) Examples() []string {
	return w.examples
}

func NewWordInfoDTO(word string, definitions []*DefinitionDTO, examples []string) *WordInfoDTO {
	return &WordInfoDTO{word: word, definitionsDTO: definitions, examples: examples}
}

func NewDefinition(definition, partOfSpeech string) *DefinitionDTO {
	return &DefinitionDTO{definition: definition, partOfSpeech: partOfSpeech}
}

type Interface interface {
	WordsInfo(ctx context.Context, words []string) ([]*WordInfoDTO, error)
}
