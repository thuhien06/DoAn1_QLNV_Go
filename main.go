package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"qlnv/database"
	"qlnv/routes"
)

type HomeData struct {
	TenDangNhap string
	VaiTro      string
	HetHan      int
	SapHetHan   int
}

// Hiển thị trang chủ theo vai trò
func home(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := routes.Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		tenDangNhap, _ := session.Values["ten_dang_nhap"].(string)
		vaiTro, _ := session.Values["vai_tro"].(string)

		// Đếm hợp đồng đã hết hạn
		var hetHan int

		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM hopdong
			WHERE ngay_ket_thuc IS NOT NULL
			  AND ngay_ket_thuc < CURDATE()
		`).Scan(&hetHan)

		if err != nil {
			http.Error(w, "Lỗi kiểm tra hợp đồng: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Đếm hợp đồng sắp hết hạn trong 30 ngày
		var sapHetHan int

		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM hopdong
			WHERE ngay_ket_thuc IS NOT NULL
			  AND ngay_ket_thuc >= CURDATE()
			  AND ngay_ket_thuc <= DATE_ADD(CURDATE(), INTERVAL 30 DAY)
		`).Scan(&sapHetHan)

		if err != nil {
			http.Error(w, "Lỗi kiểm tra hợp đồng: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := HomeData{
			TenDangNhap: tenDangNhap,
			VaiTro:      vaiTro,
			HetHan:      hetHan,
			SapHetHan:   sapHetHan,
		}

		tmpl := template.Must(
			template.ParseFiles("templates/index.html"),
		)

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
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
	http.HandleFunc("/", routes.DaDangNhap(home(db)))

	http.HandleFunc("/dangnhap", routes.DangNhap(db))
	http.HandleFunc("/dangxuat", routes.DangXuat)

	//////////////////NHÂN VIÊN

	http.HandleFunc("/quanly-nhanvien",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.QuanLyNhanVien(),
			),
		),
	)

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

	// nhân viên chỉ xem được thông tin của bản thân
	http.HandleFunc(
		"/thongtin-ca-nhan",
		routes.DaDangNhap(
			routes.CoVaiTro("nhanvien")(
				routes.ThongTinCaNhan(db),
			),
		),
	)

	http.HandleFunc(
		"/nhanvien/thongtin",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.ThongTinNhanVien(db),
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

	// Nhân viên chỉ xem được hợp đồng của bản thân
	http.HandleFunc(
		"/hopdong-ca-nhan",
		routes.DaDangNhap(
			routes.CoVaiTro("nhanvien")(
				routes.HopDongCaNhan(db),
			),
		),
	)

	http.HandleFunc("/hopdong/chi-tiet",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr", "nhanvien")(
				routes.ChiTietHopDong(db),
			),
		),
	)

	//////////////////QUẢN LÝ CÔNG

	http.HandleFunc("/cong",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.QuanLyCong(),
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

	http.HandleFunc("/chamcong/vao",
		routes.DaDangNhap(
			routes.CoVaiTro("nhanvien")(routes.ChamCongVao(db)),
		))

	http.HandleFunc("/chamcong/ra",
		routes.DaDangNhap(
			routes.CoVaiTro("nhanvien")(routes.ChamCongRa(db)),
		))

	// nhân viên chỉ xem được công của bản thâ
	http.HandleFunc(
		"/chamcong-ca-nhan",
		routes.DaDangNhap(
			routes.CoVaiTro("nhanvien")(
				routes.ChamCongCaNhan(db),
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
			routes.CoVaiTro("admin", "hr", "nhanvien")(
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

	// nhân viên chỉ xem được nghỉ phép của bản thân
	http.HandleFunc(
		"/nghiphep-ca-nhan",
		routes.DaDangNhap(
			routes.CoVaiTro("nhanvien")(
				routes.NghiPhepCaNhan(db),
			),
		),
	)

	http.HandleFunc("/nghiphep/duyet",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.XuLyDuyetNghiPhep(db),
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

	http.HandleFunc("/bangluong/luu",
		routes.DaDangNhap(
			routes.CoVaiTro("admin", "hr")(
				routes.LuuBangLuong(db),
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

	//////////////////QUAN LY TAI KHOAN- QUYỀN CỦA ADMIN

	http.HandleFunc(
		"/taikhoan",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.DanhSachTaiKhoan(db),
			),
		),
	)

	http.HandleFunc(
		"/taikhoan/them",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.ThemTaiKhoan(db),
			),
		),
	)

	http.HandleFunc(
		"/taikhoan/sua",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.SuaTaiKhoan(db),
			),
		),
	)

	http.HandleFunc(
		"/taikhoan/xoa",
		routes.DaDangNhap(
			routes.CoVaiTro("admin")(
				routes.XoaTaiKhoan(db),
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
