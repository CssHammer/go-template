package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib" // nolint
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/CssHammer/go-template/models"
)

const ServiceName = "postgres"

type Storage struct {
	db  *sqlx.DB
	log *zap.Logger
}

type Config struct {
	DSN          string
	MaxOpenConns int
}

func New(ctx context.Context, wg *sync.WaitGroup, log *zap.Logger, c Config) (*Storage, error) {
	log = log.Named(ServiceName)

	db, err := sqlx.Connect("pgx", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres connect (%s): %w", c.DSN, err)
	}

	db.SetMaxOpenConns(c.MaxOpenConns)

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := db.Close()
		if err != nil {
			log.Error("close connection:", zap.Error(err))
			return
		}
		log.Info("close connection")
	}()

	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func (s *Storage) HealthCheck() error {
	err := s.db.Ping()
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

func (s *Storage) Init(ctx context.Context) error {
	const q = `CREATE TABLE IF NOT EXISTS users
(
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(50) NOT NULL
)`
	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, id int) (*models.User, error) {
	const q = "SELECT * FROM users WHERE id = $1"

	var user models.User
	err := s.db.QueryRowxContext(ctx, q, id).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query row (id: %d): %w", id, err)
	}

	return &user, nil
}

func (s *Storage) CreateUser(ctx context.Context, user models.User) error {
	const q = `INSERT INTO users (name) values ($1)`

	_, err := s.db.ExecContext(ctx, q, user.Name)
	if err != nil {
		return fmt.Errorf("exec context (user: %v): %w", user, err)
	}

	return nil
}
