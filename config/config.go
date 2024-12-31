package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config interface {
	url() string
}

type DatabaseConfig struct {
	username string
	password string
	host     string
	port     int
	database string
}
type ServerConfig struct {
	host string
	port int
}

var database *DatabaseConfig
var server *ServerConfig

func init() {
	environment := os.Getenv("APP_ENV")
	if environment != "production" {
		environment = "development"
		loadEnvFile()

		database = &DatabaseConfig{
			username: os.Getenv("DB_USERNAME"),
			password: os.Getenv("DB_PASSWORD"),
			host:     os.Getenv("DB_HOST"),
			port:     getEnvAsInt("DB_PORT"),
			database: os.Getenv("DB_NAME"),
		}
	}
	server = &ServerConfig{
		host: os.Getenv("SERVER_HOST"),
		port: getEnvAsInt("SERVER_PORT"),
	}
}

func UrlDatabase() string {
	return database.url()
}

func UrlServer() string {
	return server.url()
}

func (database *DatabaseConfig) url() string {
	url := os.Getenv("DATABASE_URL")
	if url != "" {
		return url
	} else {
		return fmt.Sprintf(
			"host=%s user=%s dbname=%s port=%d sslmode=disable",
			database.host,
			database.username,
			database.database,
			// database.password,
			database.port,
		)
	}
}

func (server *ServerConfig) url() string {
	return fmt.Sprintf(
		"%s:%d",
		server.host,
		server.port,
	)
}

func getEnvAsInt(key string) int {
	valueStr := os.Getenv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(err)
	}
	return value
}

func loadEnvFile() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}
