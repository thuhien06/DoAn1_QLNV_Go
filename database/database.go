package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() (*sql.DB, error) {
	dsn := "root:@tcp(127.0.0.1:3306)/quanlynhanvien?parseTime=true"

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	fmt.Println("Ket noi MySQL thanh cong!")

	return db, nil
}
