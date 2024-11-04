package translator

type Translator interface {
	// Translate a text from a language to another language
	Translate(from, to Language, text string) (string, error)
	// Detects the language of the text
	Detect(text string) (Language, error)
}

type MockTranslator struct{}

func NewMockTranslator() MockTranslator {
	return MockTranslator{}
}

func (t MockTranslator) Translate(from, to Language, text string) (string, error) {
	return text, nil
}

func (t MockTranslator) Detect(text string) (Language, error) {
	return EN, nil
}
