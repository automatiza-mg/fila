package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/automatiza-mg/fila/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	image = "postgres:16-alpine"

	testUser = "test"
	testPass = "test-pw"
	testDB   = "test-db"
)

type TestInstance struct {
	pool      *dockertest.Pool
	container *dockertest.Resource
	connURL   *url.URL

	skipReason string

	mu sync.Mutex
	db *sql.DB
}

// MustTestInstance chama a função NewTestInstance e, caso ocorra algum erro, loga e encerra o processo
// com código 1 (erro).
func MustTestInstance() *TestInstance {
	ti, err := NewTestInstance()
	if err != nil {
		log.Fatal(err)
	}
	return ti
}

// NewTestInstance cria uma nova instância do banco de dados baseado em um container do Docker. Cria
// também um banco de dados inicial, executa as migrações e define o banco como template para ser
// clonado em testes futuros
//
// Essa função NÃO DEVE ser usada fora do ambiente de testes, mas é pública para ser compartilhado
// com outros packages, devendo ser chamado e instanciado na função TestMain.
//
// Todas os testes com banco de dados podem ser ignorados usando a flag -short (go test -short).
func NewTestInstance() (*TestInstance, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if testing.Short() {
		return &TestInstance{
			skipReason: "Ignorando testes de banco de dados (-short)",
		}, nil
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	repo, tag, ok := strings.Cut(image, ":")
	if !ok {
		return nil, fmt.Errorf("invalid docker image: %q", image)
	}

	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: repo,
		Tag:        tag,
		Env: []string{
			"POSTGRES_DB=" + testDB,
			"POSTGRES_USER=" + testUser,
			"POSTGRES_PASSWORD=" + testPass,
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Define o tempo de expiração do container em 3 minutos.
	if err := container.Expire(180); err != nil {
		return nil, fmt.Errorf("failed to expire container: %w", err)
	}

	connURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(testUser, testPass),
		Host:     container.GetHostPort("5432/tcp"),
		Path:     testDB,
		RawQuery: "sslmode=disable",
	}

	// Conecta com o banco de dados
	var db *sql.DB
	err = pool.Retry(func() error {
		var err error
		db, err = sql.Open("pgx", connURL.String())
		if err != nil {
			return err
		}
		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(1)

		if err := db.Ping(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Executa as migrações
	if err := migrations.Up(db); err != nil {
		return nil, err
	}

	return &TestInstance{
		pool:      pool,
		container: container,
		connURL:   connURL,
		db:        db,
	}, nil
}

// Close remove os recursos utilizados pelo Docker e fecha a conexão com o banco de dados.
func (i *TestInstance) Close() error {
	if i.skipReason != "" {
		return nil
	}

	errs := make([]error, 0)
	if err := i.pool.Purge(i.container); err != nil {
		errs = append(errs, err)
	}

	err := i.db.Close()
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// NewDatabase cria um novo banco de dados com base no template para uso em testes.
func (i *TestInstance) NewDatabase(tb testing.TB) *pgxpool.Pool {
	tb.Helper()

	if i.skipReason != "" {
		tb.Skip(i.skipReason)
	}

	dbName, err := i.clone()
	if err != nil {
		tb.Fatalf("failed to clone database: %v", err)
	}

	connURL := i.connURL.ResolveReference(&url.URL{Path: dbName})
	connURL.RawQuery = "sslmode=disable"

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connURL.String())
	if err != nil {
		tb.Fatalf("failed to connect to database %q: %v", dbName, err)
	}

	tb.Cleanup(func() {
		pool.Close()

		i.mu.Lock()
		defer i.mu.Unlock()

		q := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE);`, dbName)
		if _, err := i.db.Exec(q); err != nil {
			tb.Errorf("failed to drop database %q: %v", dbName, err)
		}
	})

	return pool
}

// Cria uma clone do banco de dados usando o template.
func (i *TestInstance) clone() (string, error) {
	dbName := rand.Text()

	q := fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE "%s"`, dbName, testDB)

	i.mu.Lock()
	defer i.mu.Unlock()

	if _, err := i.db.Exec(q); err != nil {
		return "", fmt.Errorf("failed to clone template database: %w", err)
	}
	return dbName, nil
}
