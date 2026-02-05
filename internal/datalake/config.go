package datalake

import (
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	User     string `env:"DATALAKE_USER,notEmpty"`
	Password string `env:"DATALAKE_PASSWORD,notEmpty" json:"-"`
	Host     string `env:"DATALAKE_HOST,notEmpty"`
	Port     int    `env:"DATALAKE_PORT,notEmpty"`
}

func (c *Config) connString() string {
	q := make(url.Values)
	q.Set("auth", "ldap")
	q.Set("tls", "true")

	u := &url.URL{
		Scheme:   "impala",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		RawQuery: q.Encode(),
	}
	return u.String()
}
