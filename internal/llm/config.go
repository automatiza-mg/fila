package llm

type Config struct {
	AzureURL    string `env:"AZURE_OPENAI_URL,notEmpty"`
	AzureApiKey string `env:"AZURE_OPENAI_API_KEY,notEmpty"`
}
