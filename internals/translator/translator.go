package translator

import "dislate/internals/lang"

type Translator interface {
	// Translate a text from a language to another language
	Translate(from, to lang.Language, text string) (string, error)
	// Detects the language of the text
	Detect(text string) (lang.Language, error)
}

type MockTranslator struct{}

func NewMockTranslator() MockTranslator {
	return MockTranslator{}
}
func (t MockTranslator) Translate(from, to lang.Language, text string) (string, error) {
	return text, nil
}
func (t MockTranslator) Detect(text string) (lang.Language, error) {
	return lang.EN, nil
}
