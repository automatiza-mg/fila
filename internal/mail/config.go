package mail

type Config struct {
	User        string `env:"MAIL_USER"`
	Password    string `env:"MAIL_PASSWORD" json:"-"`
	Host        string `env:"MAIL_HOST"`
	Port        int    `env:"MAIL_PORT"`
	FromAddress string `env:"MAIL_FROM_ADDRESS,notEmpty" envDefault:"notificacoes@planejamento.mg.gov.br"`
}
