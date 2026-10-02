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
	http.HandleFunc("/dangxuat", routes.DangXuat)

	//////////////////NHÂN VIÊN

	http.HandleFunc("/nhanvien",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.DanhSachNhanVien(db),
			),
		),
	)

	http.HandleFunc("/nhanvien/them",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.ThemNhanVien(db),
			),
		),
	)

	http.HandleFunc("/nhanvien/sua",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.SuaNhanVien(db),
			),
		),
	)

	http.HandleFunc("/nhanvien/xoa",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.XoaNhanVien(db),
			),
		),
	)

	//////////////////HỢP ĐỒNG

	// Khai báo đường dẫn quản lý hợp đồng
	http.HandleFunc("/hopdong",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.DanhSachHopDong(db),
			),
		),
	)

	http.HandleFunc("/hopdong/them",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.ThemHopDong(db),
			),
		),
	)

	http.HandleFunc("/hopdong/sua",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.SuaHopDong(db),
			),
		),
	)

	http.HandleFunc("/hopdong/xoa",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.XoaHopDong(db),
			),
		),
	)
	//////////////////CHẤM CÔNG

	// Khai báo đường dẫn quản lý chấm công
	http.HandleFunc("/chamcong",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.DanhSachChamCong(db),
			),
		),
	)

	http.HandleFunc("/chamcong/them",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.ThemChamCong(db),
			),
		),
	)

	http.HandleFunc("/chamcong/sua",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.SuaChamCong(db),
			),
		),
	)

	http.HandleFunc("/chamcong/xoa",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.XoaChamCong(db),
			),
		),
	)

	//////////////////NGHI PHEP

	http.HandleFunc("/nghiphep",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.DanhSachNghiPhep(db),
			),
		),
	)

	http.HandleFunc("/nghiphep/them",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.ThemNghiPhep(db),
			),
		),
	)

	http.HandleFunc("/nghiphep/sua",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.SuaNghiPhep(db),
			),
		),
	)

	http.HandleFunc("/nghiphep/xoa",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.XoaNghiPhep(db),
			),
		),
	)

	//////////////////BANG LUONG

	http.HandleFunc("/bangluong",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.DanhSachBangLuong(db),
			),
		),
	)

	http.HandleFunc("/bangluong/them",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.ThemBangLuong(db),
			),
		),
	)

	http.HandleFunc("/bangluong/sua",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.SuaBangLuong(db),
			),
		),
	)

	http.HandleFunc("/bangluong/xoa",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.XoaBangLuong(db),
			),
		),
	)

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
