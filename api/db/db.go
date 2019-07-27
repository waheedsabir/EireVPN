package db

import (
	"eirevpn/api/models"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

var db *gorm.DB
var err error

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Init creates a connection to postgres database and
// migrates any new models
func Init(debug bool) {
	userDb := getEnv("PG_USER", "")
	password := getEnv("PG_PASSWORD", "")
	host := getEnv("PG_HOST", "")
	port := getEnv("PG_PORT", "5432")
	database := getEnv("PG_DB", "")

	dbinfo := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		userDb,
		password,
		host,
		port,
		database,
	)

	fmt.Println(dbinfo)

	db, err = gorm.Open("postgres", dbinfo)
	db.LogMode(debug)

	if err != nil {
		log.Println("Failed to connect to database")
		panic(err)
	}
	log.Println("Database connected")
	if !db.HasTable(&models.User{}) {
		err := db.CreateTable(&models.User{})
		if err == nil {
			log.Println("Table Created")
		}
	}

	if !db.HasTable(&models.Plan{}) {
		err := db.CreateTable(&models.Plan{})
		if err == nil {
			log.Println("Table Created")
		}
	}

	if !db.HasTable(&models.UserPlan{}) {
		err := db.CreateTable(&models.UserPlan{})
		if err == nil {
			log.Println("Table Created")
		} 
	}

	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Plan{})
	db.AutoMigrate(&models.UserPlan{})
}

//GetDB ...
func GetDB() *gorm.DB {
	return db
}

func CloseDB() {
	db.Close()
}
