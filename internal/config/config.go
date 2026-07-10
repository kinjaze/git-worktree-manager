package config

const (
	DefaultLanguage = "en"
)

type Config struct {
	Language string `json:"language"`
}

func Default() Config {
	return Config{Language: DefaultLanguage}
}

func NormalizeLanguage(language string) string {
	switch language {
	case "zh":
		return "zh"
	case "en", "":
		return "en"
	default:
		return "en"
	}
}
