package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.syneforge.com/gin/go-common-lb/mongo"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"

	"github.com/CssHammer/go-template/models"
)

const (
	ServiceName = "mongo"

	CollectionUsers = "users"
)

type Storage struct {
	log     *zap.Logger
	manager *mongo.SessionManager
}

type Config struct {
	DSN    string
	DBName string
}

func New(ctx context.Context, wg *sync.WaitGroup, log *zap.Logger, c Config) (*Storage, error) {
	log = log.Named(ServiceName)

	manager, err := mongo.NewSessionBuilder().
		SetRequiredContext(ctx).
		SetRequiredDSN(c.DSN).
		SetRequiredDBName(c.DBName).
		SetRequiredLogger(log).
		Build()
	if err != nil {
		return nil, fmt.Errorf("init mongo session manager: %w", err)
	}

	err = manager.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		manager.Close(shutdownCtx)
		log.Info("close connection")
	}()

	return &Storage{
		log:     log,
		manager: manager,
	}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	_, err := s.manager.Collection(CollectionUsers).Indexes().CreateOne(
		ctx,
		mongoDriver.IndexModel{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}

	return nil
}

func (s *Storage) HealthCheck() error {
	return nil
}

func (s *Storage) GetUser(ctx context.Context, id int) (*models.User, error) {
	var item models.User

	result := s.manager.Collection(CollectionUsers).FindOne(ctx, s.newGetUserFilter(id))
	err := result.Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("decode user (id: %d): %w", id, err)
	}

	return &item, nil
}

func (s *Storage) CreateUser(ctx context.Context, user models.User) error {
	_, err := s.manager.Collection(CollectionUsers).InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("create user (user: %v): %w", user, err)
	}

	return nil
}

func (s *Storage) newGetUserFilter(id int) interface{} {
	return bson.M{"id": id}
}
