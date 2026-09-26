package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Connect kết nối đến cơ sở dữ liệu MySQL
func Connect() (*sql.DB, error) {

	// Thông tin kết nối đến database
	dsn := "root:@tcp(127.0.0.1:3306)/quanlynhanvien?parseTime=true"

	// Mở kết nối MySQL
	db, err := sql.Open("mysql", dsn)

	// Kiểm tra lỗi khi mở kết nối
	if err != nil {
		return nil, err
	}

	// Kiểm tra kết nối đến MySQL
	err = db.Ping()

	// Nếu kết nối thất bại thì trả về lỗi
	if err != nil {
		return nil, err
	}

	// Thông báo kết nối thành công
	fmt.Println("Kết nối MySQL thành công!")

	// Trả về kết nối database
	return db, nil
}
