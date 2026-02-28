package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	Port             string
	MongoURI         string
	MongoDatabase    string
	JWTSecret        string
	JWTExpiryHours   int
	CSRFSecret       string
	RateLimitReq     int
	RateLimitWindow  int
}

var (
	cfg  *Config
	once sync.Once
)

func Load() *Config {
	once.Do(func() {
		cfg = &Config{
			Port:            getEnv("PORT", "8080"),
			MongoURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			MongoDatabase:   getEnv("MONGODB_DATABASE", "leadflowcrm"),
			JWTSecret:       getEnv("JWT_SECRET", "default_secret"),
			JWTExpiryHours:  24,
			CSRFSecret:      getEnv("CSRF_SECRET", "csrf_secret"),
			RateLimitReq:    100,
			RateLimitWindow: 1,
		}
	})
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if err := loadEnvFile(); err == nil {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}

func loadEnvFile() error {
	return LoadEnvFile(".env")
}

func LoadEnvFile(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	for _, line := range contents {
		lineStr := string(line)
		if len(lineStr) > 0 && lineStr[0] == '#' {
			continue
		}
		for i := 0; i < len(lineStr); i++ {
			if lineStr[i] == '=' {
				key := lineStr[:i]
				value := lineStr[i+1:]
				os.Setenv(key, value)
				break
			}
		}
	}
	return nil
}

type Database struct {
	Client   *mongo.Client
	Database *mongo.Database
}

var db *Database
var dbOnce sync.Once

func ConnectDB() (*Database, error) {
	var err error
	dbOnce.Do(func() {
		cfg := Load()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOpts := options.Client().ApplyURI(cfg.MongoURI)
		client, err := mongo.Connect(ctx, clientOpts)
		if err != nil {
			err = fmt.Errorf("failed to connect to MongoDB: %w", err)
			return
		}

		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			err = fmt.Errorf("failed to ping MongoDB: %w", err)
			return
		}

		log.Println("Connected to MongoDB successfully")
		db = &Database{
			Client:   client,
			Database: client.Database(cfg.MongoDatabase),
		}
	})
	return db, err
}

func GetDatabase() *Database {
	return db
}
