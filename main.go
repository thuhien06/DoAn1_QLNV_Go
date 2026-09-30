package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"qlnv/database"
	"qlnv/routes"
)

// Hiển thị trang chủ
func home(w http.ResponseWriter, r *http.Request) {
	// Đọc file giao diện trang chủ
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	// Hiển thị giao diện
	tmpl.Execute(w, nil)
}

func main() {

	// Kết nối đến cơ sở dữ liệu MySQL
	db, err := database.Connect()

	// Kiểm tra lỗi kết nối
	if err != nil {
		fmt.Println("Lỗi kết nối database:", err)
		return
	}

	// Đóng kết nối database khi chương trình kết thúc
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	// Khai báo đường dẫn trang chủ
	http.HandleFunc("/", home)

	//////////////////NHÂN VIÊN

	// Khai báo đường dẫn quản lý nhân viên
	http.HandleFunc("/nhanvien", routes.DanhSachNhanVien(db))

	// Thêm nhân viên
	http.HandleFunc("/nhanvien/them", routes.ThemNhanVien(db))

	// Khai báo đường dẫn sửa nhân viên
	http.HandleFunc("/nhanvien/sua", routes.SuaNhanVien(db))

	// Xóa nhân viên
	http.HandleFunc("/nhanvien/xoa", routes.XoaNhanVien(db))

	//////////////////HỢP ĐỒNG

	// Khai báo đường dẫn quản lý hợp đồng
	http.HandleFunc("/hopdong", routes.DanhSachHopDong(db))

	// Khai báo đường dẫn thêm hợp đồng
	http.HandleFunc("/hopdong/them", routes.ThemHopDong(db))

	// Khai báo đường dẫn sửa hợp đồng
	http.HandleFunc("/hopdong/sua", routes.SuaHopDong(db))

	// Khai báo đường dẫn xóa hợp đồng
	http.HandleFunc("/hopdong/xoa", routes.XoaHopDong(db))

	//////////////////CHẤM CÔNG

	// Khai báo đường dẫn quản lý chấm công
	http.HandleFunc("/chamcong", routes.DanhSachChamCong(db))

	// Khai báo đường dẫn thêm chấm công
	http.HandleFunc("/chamcong/them", routes.ThemChamCong(db))

	// Khai báo đường dẫn sửa chấm công
	http.HandleFunc("/chamcong/sua", routes.SuaChamCong(db))

	// Khai báo đường dẫn xóa chấm công
	http.HandleFunc("/chamcong/xoa", routes.XoaChamCong(db))

	// Cho phép truy cập các file CSS, JavaScript, hình ảnh...
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("static")),
		),
	)

	// Thông báo địa chỉ server
	fmt.Println("Server đang chạy tại http://localhost:8080")

	// Khởi động web server tại cổng 8080
	err = http.ListenAndServe(":8080", nil)

	// Kiểm tra lỗi khi khởi động server
	if err != nil {
		fmt.Println("Lỗi:", err)
	}
}
