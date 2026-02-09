package blob

type Config struct {
	// Define o tipo de storage que será utilizado. Os valores possíveis são 'filesystem' e 'azure'.
	Provider string `env:"STORAGE_PROVIDER,notEmpty" default:"filesystem"`
	// O diretório raiz que será usado pela FilesystemStore. O valor padrão utilizado é '.blob'.
	FilesystemRoot string `env:"STORAGE_FILESYSTEM_ROOT" default:".blob"`
	AzureAccount   string `env:"STORAGE_AZURE_ACCOUNT"`
	AzureContainer string `env:"STORAGE_AZURE_CONTAINER"`
	AzureApiKey    string `env:"STORAGE_AZURE_API_KEY"`
}
