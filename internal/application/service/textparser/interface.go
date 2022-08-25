package textparser

type Interface interface {
	ExtractWords(text string) ([]string, error)
}
