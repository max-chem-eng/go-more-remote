package models

import (
	"fmt"

	"github.com/max-chem-eng/go-more-remote/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func init() {
	CreateConnection()
}

func CreateConnection() {
	if GetConnection() != nil {
		return
	}

	url := config.UrlDatabase()
	if connection, err := gorm.Open(postgres.Open(url)); err != nil {
		panic(err)
	} else {
		Db = connection
		fmt.Println("Connection established")
	}
}

func CloseConnection() {
	sqlDB, _ := Db.DB()
	sqlDB.Close()
	fmt.Println("Connection closed")
}

func CreateTables() {
	Db.AutoMigrate(&Job{}, &JobRun{}, &Attachment{})
	// migrator := Db.Migrator()
	// if !migrator.HasTable(&Job{}) {
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("Job table created")
	// } else {
	// 	fmt.Println("Job table already exists")
	// }
}

func GetConnection() *gorm.DB {
	return Db
}
