package database

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var RClient *redis.Client
var Ctx = context.Background()

func Connect(db_url string, redis_url string) {
	dsn := db_url

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Database error:", err.Error()) //It should be Panic
	}

	fmt.Println("🚀 Connected to PostgreSQL successfully!")

	opt, err := redis.ParseURL("redis://<user>:<pass>@localhost:6379/<db>")
	if err != nil {
		panic(err)
	}

	RClient = redis.NewClient(opt)

	fmt.Println("🚀 Connected to Redis successfully!")

}
