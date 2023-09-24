package storage

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mongodb/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BootstrapMongo(uri string, dbName string, timeout time.Duration) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	return client.Database(dbName), nil
}

func StartSessionManager(uri string) *session.Store {
	mongoStorage := mongodb.New(
		mongodb.Config{
			ConnectionURI: uri,
			Database:      "test",
			Collection:    "session_storage",
			Reset:         false,
		},
	)
	return session.New(
		session.Config{
			Expiration:     60 * 60 * 24 * 30 * 1000,
			Storage:        mongoStorage,
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: "strict",
		},
	)
}

func CloseMongo(db *mongo.Database) error {
	return db.Client().Disconnect(context.Background())
}
