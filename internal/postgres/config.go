package postgres

import (
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	User     string `env:"POSTGRES_USER,notEmpty"`
	Password string `env:"POSTGRES_PASSWORD,notEmpty" json:"-"`
	Host     string `env:"POSTGRES_HOST,notEmpty"`
	Port     int    `env:"POSTGRES_PORT,notEmpty"`
	DB       string `env:"POSTGRES_DB,notEmpty"`
}

func (c *Config) connString() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.DB,
	}
	return u.String()
}
