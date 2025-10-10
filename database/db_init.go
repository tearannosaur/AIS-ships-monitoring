package database

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func DbInit() error {
	//dsn := "root:ais@tcp(127.0.0.1:1433)/ais_database?parseTime=true&charset=utf8mb4"
	//dsn := "sqlserver://sa:master@127.0.0.1:1433?database=ais_database&encrypt=disable"
	dsn := "sqlserver://@localhost?database=ais_database&trusted_connection=yes"
	var err error
	DB, err = sqlx.Connect("sqlserver", dsn)
	if err != nil {
		return err

	}
	return nil
}

func DdClose() {
	fmt.Println("Базы данных выключилась")
	DB.Close()
}
