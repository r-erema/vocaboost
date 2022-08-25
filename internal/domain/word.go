package domain

type Word struct {
	word string
	definitions,
	examples,
	imageURLs []string
}

func (w *Word) Definitions() []string {
	return w.definitions
}

func (w *Word) Examples() []string {
	return w.examples
}

func (w *Word) ImageURLs() []string {
	return w.imageURLs
}

func NewWord(word string, definitions, examples, imageURLs []string) *Word {
	return &Word{word: word, definitions: definitions, examples: examples, imageURLs: imageURLs}
}

func (w *Word) Word() string {
	return w.word
}
