package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go-geom-basics/facilities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host     = "localhost"
	port     = 6432
	user     = "postgres"
	password = "postgres"
	dbname   = "go-geom"
)

func main() {
	fmt.Println("Program Initialization")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	sqlDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {

		}
	}(sqlDB)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	facilities.CreateDB(gormDB)
	facilities.AddMedicalFacilities(gormDB)
	fmt.Println("***...Exiting....***")
}
