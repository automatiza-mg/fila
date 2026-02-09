package docintel

type Config struct {
	AzureURL    string `env:"AZURE_DOC_URL,notEmpty"`
	AzureApiKey string `env:"AZURE_DOC_API_KEY,notEmpty"`
}
