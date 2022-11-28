package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host = "localhost"
	port = 5432
	// user = "postgres"
	// password = "your-password"
	dbname = "learn_sql"
)

type config struct {
	port int
	// env    string
	host   string
	dbname string
}

func main() {

	config := config{
		port:   5432,
		host:   "localhost",
		dbname: "learn_sql",
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable",
		config.host, config.port, config.dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")

	company, err := assignCompany("msft")
	if err != nil {
		panic(err)
	}

	// company.CreateCompanyTable(db)
	company.CreateCompanyFinancials(db)

}
