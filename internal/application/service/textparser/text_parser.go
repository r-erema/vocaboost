package textparser

import (
	"regexp"
	"strings"
)

type V1 struct{}

func (tp V1) ExtractWords(text string) ([]string, error) {
	re := regexp.MustCompile(`\W`)
	text = re.ReplaceAllString(text, " ")

	return strings.Fields(text), nil
}
