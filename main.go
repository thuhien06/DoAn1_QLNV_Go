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

	// Khai báo đường đăng nhập  rồi vào trang chủ
	http.HandleFunc("/", routes.DaDangNhap(home))

	http.HandleFunc("/dangnhap", routes.DangNhap(db))

	//////////////////NHÂN VIÊN

	// Khai báo đường dẫn quản lý nhân viên sau đăng nhập
	http.HandleFunc("/nhanvien", routes.DaDangNhap(routes.DanhSachNhanVien(db)))

	// Thêm nhân viên
	http.HandleFunc("/nhanvien/them", routes.DaDangNhap(routes.ThemNhanVien(db)))

	// Khai báo đường dẫn sửa nhân viên
	http.HandleFunc("/nhanvien/sua", routes.DaDangNhap(routes.SuaNhanVien(db)))

	// Xóa nhân viên
	http.HandleFunc("/nhanvien/xoa", routes.DaDangNhap(routes.XoaNhanVien(db)))

	//////////////////HỢP ĐỒNG

	// Khai báo đường dẫn quản lý hợp đồng
	http.HandleFunc("/hopdong", routes.DaDangNhap(routes.DanhSachHopDong(db)))
	http.HandleFunc("/hopdong/them", routes.DaDangNhap(routes.ThemHopDong(db)))
	http.HandleFunc("/hopdong/sua", routes.DaDangNhap(routes.SuaHopDong(db)))
	http.HandleFunc("/hopdong/xoa", routes.DaDangNhap(routes.XoaHopDong(db)))

	//////////////////CHẤM CÔNG

	// Khai báo đường dẫn quản lý chấm công
	http.HandleFunc("/chamcong", routes.DaDangNhap(routes.DanhSachChamCong(db)))
	http.HandleFunc("/chamcong/them", routes.DaDangNhap(routes.ThemChamCong(db)))
	http.HandleFunc("/chamcong/sua", routes.DaDangNhap(routes.SuaChamCong(db)))
	http.HandleFunc("/chamcong/xoa", routes.DaDangNhap(routes.XoaChamCong(db)))

	//////////////////NGHI PHEP

	// Khai báo đường dẫn quản lý nghỉ phép
	http.HandleFunc("/nghiphep", routes.DaDangNhap(routes.DanhSachNghiPhep(db)))
	http.HandleFunc("/nghiphep/them", routes.DaDangNhap(routes.ThemNghiPhep(db)))
	http.HandleFunc("/nghiphep/sua", routes.DaDangNhap(routes.SuaNghiPhep(db)))
	http.HandleFunc("/nghiphep/xoa", routes.DaDangNhap(routes.XoaNghiPhep(db)))

	//////////////////BANG LUONG

	// Khai báo đường dẫn quản lý bảng lương
	http.HandleFunc("/bangluong", routes.DaDangNhap(routes.DanhSachBangLuong(db)))
	http.HandleFunc("/bangluong/them", routes.DaDangNhap(routes.ThemBangLuong(db)))
	http.HandleFunc("/bangluong/sua", routes.DaDangNhap(routes.SuaBangLuong(db)))
	http.HandleFunc("/bangluong/xoa", routes.DaDangNhap(routes.XoaBangLuong(db)))

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
