package i18n

import "fmt"

type Translator struct {
	language string
	messages map[string]string
}

func New(language string) Translator {
	if language == "zh" {
		return Translator{language: "zh", messages: zhMessages}
	}
	return Translator{language: "en", messages: enMessages}
}

func (t Translator) Language() string {
	return t.language
}

func (t Translator) T(key string, args ...any) string {
	message, ok := t.messages[key]
	if !ok {
		message = key
	}
	if len(args) == 0 {
		return message
	}
	return fmt.Sprintf(message, args...)
}
