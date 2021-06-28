package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go-geom-basics/facilities"
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
	//gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	//if err != nil {
	//	panic(err)
	//}

	facilities.CreateDB(sqlDB)
	facilities.AddMedicalFacilities(sqlDB)
	fmt.Println("***...Exiting....***")
}
