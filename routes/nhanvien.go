package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"qlnv/models"
)

// struct của quản lý nhân viên
type ThongTinHopDong struct {
	MaHD          string
	LoaiHopDong   string
	NhomHopDong   string
	NgayBatDau    *time.Time
	NgayKetThuc   *time.Time
	LuongCoBan    float64
	GhiChu        string
	DoiTac        string
	NguoiDaiDien  string
	NoiDung       string
	GiaTri        float64
	NgayKy        *time.Time
	DichVuThietBi string
	ChiPhi        float64
}

type ThongTinChamCong struct {
	NgayChamCong time.Time
	GioVao       string
	GioRa        string
	TrangThai    string
	GhiChu       string
}

type ThongTinNghiPhep struct {
	NgayBatDau  time.Time
	NgayKetThuc time.Time
	LoaiNghi    string
	LyDo        string
	TrangThai   string
	GhiChu      string
}

type ThongTinBangLuong struct {
	Thang      int
	Nam        int
	LuongCoBan float64
	SoNgayCong int
	PhuCap     float64
	KhauTru    float64
	TongLuong  float64
	GhiChu     string
}

// Hiển thị danh sách nhân viên
func DanhSachNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maNV := r.URL.Query().Get("ma_nv")

		// Lấy tổng số nhân viên trong hệ thống
		var tongSo int
		err := db.QueryRow("SELECT COUNT(*) FROM nhanvien").Scan(&tongSo)
		if err != nil {
			http.Error(w, "Lỗi đếm số nhân viên", http.StatusInternalServerError)
			return
		}
		// Lấy danh sách nhân viên
		var rows *sql.Rows

		if maNV != "" {
			rows, err = db.Query(`
				SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
				       so_dien_thoai, email, dia_chi, ngay_vao_lam,
				       phong_ban, chuc_vu
				FROM nhanvien
				WHERE ma_nv LIKE ?
				ORDER BY id
			`, "%"+maNV+"%")
		} else {
			rows, err = db.Query(`
				SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
				       so_dien_thoai, email, dia_chi, ngay_vao_lam,
				       phong_ban, chuc_vu
				FROM nhanvien
				ORDER BY id
			`)
		}

		if err != nil {
			http.Error(w, "Lỗi truy vấn database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.NhanVien

		for rows.Next() {
			var nv models.NhanVien

			err := rows.Scan(
				&nv.ID,
				&nv.MaNV,
				&nv.HoTen,
				&nv.NgaySinh,
				&nv.GioiTinh,
				&nv.SoDienThoai,
				&nv.Email,
				&nv.DiaChi,
				&nv.NgayVaoLam,
				&nv.PhongBan,
				&nv.ChucVu,
			)
			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu", http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, nv)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc danh sách nhân viên", http.StatusInternalServerError)
			return
		}

		// Dữ liệu truyền sang giao diện
		data := struct {
			TongSo   int
			TuKhoa   string
			DanhSach []models.NhanVien
		}{
			TongSo:   len(danhSach),
			TuKhoa:   maNV,
			DanhSach: danhSach,
		}

		tmpl := template.Must(template.ParseFiles(
			"templates/nhanvien/list.html",
		))

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị danh sách", http.StatusInternalServerError)
		}
	}
}

// Thêm nhân viên
func ThemNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Nếu là GET thì hiển thị form thêm nhân viên
		if r.Method == http.MethodGet {

			tmpl := template.Must(template.ParseFiles(
				"templates/nhanvien/them.html",
			))

			tmpl.Execute(w, nil)
			return
		}

		// Lấy dữ liệu từ form
		maNV := r.FormValue("ma_nv")
		hoTen := r.FormValue("ho_ten")
		ngaySinh := r.FormValue("ngay_sinh")
		gioiTinh := r.FormValue("gioi_tinh")
		soDienThoai := r.FormValue("so_dien_thoai")
		email := r.FormValue("email")
		diaChi := r.FormValue("dia_chi")
		ngayVaoLam := r.FormValue("ngay_vao_lam")
		phongBan := r.FormValue("phong_ban")
		chucVu := r.FormValue("chuc_vu")

		// Thêm nhân viên vào database
		_, err := db.Exec(`
			INSERT INTO nhanvien
			(ma_nv, ho_ten, ngay_sinh, gioi_tinh, so_dien_thoai,
			 email, dia_chi, ngay_vao_lam, phong_ban, chuc_vu)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			maNV,
			hoTen,
			ngaySinh,
			gioiTinh,
			soDienThoai,
			email,
			diaChi,
			ngayVaoLam,
			phongBan,
			chucVu,
		)

		// Kiểm tra lỗi
		if err != nil {
			http.Error(w, "Lỗi thêm nhân viên: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Thêm thành công thì quay về danh sách
		http.Redirect(w, r, "/nhanvien", http.StatusSeeOther)
	}
}

// Sửa thông tin nhân viên
func SuaNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Lấy ID nhân viên từ URL
		id := r.URL.Query().Get("id")

		// Nếu là GET thì hiển thị thông tin hiện tại
		if r.Method == http.MethodGet {

			var nv models.NhanVien

			// Lấy thông tin nhân viên từ database
			err := db.QueryRow(`
				SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
				       so_dien_thoai, email, dia_chi, ngay_vao_lam,
				       phong_ban, chuc_vu
				FROM nhanvien
				WHERE id = ?
			`, id).Scan(
				&nv.ID,
				&nv.MaNV,
				&nv.HoTen,
				&nv.NgaySinh,
				&nv.GioiTinh,
				&nv.SoDienThoai,
				&nv.Email,
				&nv.DiaChi,
				&nv.NgayVaoLam,
				&nv.PhongBan,
				&nv.ChucVu,
			)

			// Kiểm tra lỗi
			if err != nil {
				http.Error(w, "Không tìm thấy nhân viên", http.StatusNotFound)
				return
			}

			// Đọc giao diện sửa
			tmpl := template.Must(template.ParseFiles(
				"templates/nhanvien/sua.html",
			))

			// Hiển thị thông tin nhân viên
			tmpl.Execute(w, nv)

			return
		}

		// Lấy dữ liệu mới từ form
		maNV := r.FormValue("ma_nv")
		hoTen := r.FormValue("ho_ten")
		ngaySinh := r.FormValue("ngay_sinh")
		gioiTinh := r.FormValue("gioi_tinh")
		soDienThoai := r.FormValue("so_dien_thoai")
		email := r.FormValue("email")
		diaChi := r.FormValue("dia_chi")
		ngayVaoLam := r.FormValue("ngay_vao_lam")
		phongBan := r.FormValue("phong_ban")
		chucVu := r.FormValue("chuc_vu")

		// Cập nhật thông tin nhân viên
		_, err := db.Exec(`
			UPDATE nhanvien
			SET ma_nv = ?,
			    ho_ten = ?,
			    ngay_sinh = ?,
			    gioi_tinh = ?,
			    so_dien_thoai = ?,
			    email = ?,
			    dia_chi = ?,
			    ngay_vao_lam = ?,
			    phong_ban = ?,
			    chuc_vu = ?
			WHERE id = ?
		`,
			maNV,
			hoTen,
			ngaySinh,
			gioiTinh,
			soDienThoai,
			email,
			diaChi,
			ngayVaoLam,
			phongBan,
			chucVu,
			id,
		)

		// Kiểm tra lỗi
		if err != nil {
			http.Error(w, "Lỗi sửa nhân viên: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Sửa thành công thì quay lại danh sách
		http.Redirect(w, r, "/nhanvien", http.StatusSeeOther)
	}
}

// Xóa nhân viên
func XoaNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Lấy ID nhân viên
		id := r.URL.Query().Get("id")

		// Chuyển ID sang số
		idNV, err := strconv.Atoi(id)

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		// Xóa nhân viên
		_, err = db.Exec(
			"DELETE FROM nhanvien WHERE id = ?",
			idNV,
		)

		if err != nil {
			http.Error(w, "Lỗi xóa nhân viên", http.StatusInternalServerError)
			return
		}

		// Quay lại danh sách
		http.Redirect(w, r, "/nhanvien", http.StatusSeeOther)
	}
}

// Thông tin cá nhân của nhân viên đang đăng nhập
func ThongTinCaNhan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		maNV, ok := session.Values["ma_nv"].(string)
		if !ok || maNV == "" {
			http.Error(w, "Tài khoản chưa được liên kết với nhân viên", http.StatusForbidden)
			return
		}

		var nv models.NhanVien

		err = db.QueryRow(`
			SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
			       so_dien_thoai, email, dia_chi,
			       ngay_vao_lam, phong_ban, chuc_vu
			FROM nhanvien
			WHERE ma_nv = ?
		`, maNV).Scan(
			&nv.ID,
			&nv.MaNV,
			&nv.HoTen,
			&nv.NgaySinh,
			&nv.GioiTinh,
			&nv.SoDienThoai,
			&nv.Email,
			&nv.DiaChi,
			&nv.NgayVaoLam,
			&nv.PhongBan,
			&nv.ChucVu,
		)

		if err != nil {
			http.Error(w, "Không tìm thấy thông tin nhân viên", http.StatusNotFound)
			return
		}

		tmpl := template.Must(
			template.ParseFiles("templates/nhanvien/thongtin.html"),
		)

		err = tmpl.Execute(w, nv)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Xem thông tin tổng hợp của một nhân viên
func ThongTinNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		maNV := r.URL.Query().Get("ma_nv")

		tmpl := template.Must(
			template.ParseFiles("templates/nhanvien/thongtin_nhanvien.html"),
		)

		// Chưa nhập mã nhân viên
		if maNV == "" {
			data := struct {
				DaTim      bool
				NhanVien   *models.NhanVien
				HopDongs   []ThongTinHopDong
				ChamCongs  []ThongTinChamCong
				NghiPheps  []ThongTinNghiPhep
				BangLuongs []ThongTinBangLuong
				Error      string
			}{
				DaTim: false,
			}

			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Lỗi hiển thị trang", http.StatusInternalServerError)
			}
			return
		}

		// =========================
		// 1. THÔNG TIN NHÂN VIÊN
		// =========================

		var nv models.NhanVien

		err := db.QueryRow(`
			SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
			       so_dien_thoai, email, dia_chi,
			       ngay_vao_lam, phong_ban, chuc_vu
			FROM nhanvien
			WHERE ma_nv = ?
		`, maNV).Scan(
			&nv.ID,
			&nv.MaNV,
			&nv.HoTen,
			&nv.NgaySinh,
			&nv.GioiTinh,
			&nv.SoDienThoai,
			&nv.Email,
			&nv.DiaChi,
			&nv.NgayVaoLam,
			&nv.PhongBan,
			&nv.ChucVu,
		)

		// Không tìm thấy nhân viên
		if err == sql.ErrNoRows {
			data := struct {
				DaTim    bool
				NhanVien *models.NhanVien
				Error    string
			}{
				DaTim: false,
				Error: "Không tìm thấy nhân viên có mã: " + maNV,
			}

			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Lỗi hiển thị trang", http.StatusInternalServerError)
			}
			return
		}

		if err != nil {
			http.Error(
				w,
				"Lỗi truy vấn thông tin nhân viên: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		// =========================
		// 2. HỢP ĐỒNG
		// =========================

		var hopDongs []ThongTinHopDong

		rows, err := db.Query(`
			SELECT ma_hd,
				COALESCE(loai_hop_dong, ''),
				COALESCE(nhom_hop_dong, ''),
				ngay_bat_dau,
				ngay_ket_thuc,
				COALESCE(luong_co_ban, 0),
				COALESCE(ghi_chu, ''),
				COALESCE(doi_tac, ''),
				COALESCE(nguoi_dai_dien, ''),
				COALESCE(noi_dung, ''),
				COALESCE(gia_tri, 0),
				ngay_ky,
				COALESCE(dich_vu_thiet_bi, ''),
				COALESCE(chi_phi, 0)
			FROM hopdong
			WHERE ma_nv = ?
			ORDER BY ngay_bat_dau DESC
		`, maNV)

		if err != nil {
			http.Error(
				w,
				"Lỗi truy vấn hợp đồng: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		for rows.Next() {
			var hd ThongTinHopDong

			err := rows.Scan(
				&hd.MaHD,
				&hd.LoaiHopDong,
				&hd.NhomHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&hd.LuongCoBan,
				&hd.GhiChu,
				&hd.DoiTac,
				&hd.NguoiDaiDien,
				&hd.NoiDung,
				&hd.GiaTri,
				&hd.NgayKy,
				&hd.DichVuThietBi,
				&hd.ChiPhi,
			)

			if err != nil {
				rows.Close()
				http.Error(
					w,
					"Lỗi đọc dữ liệu hợp đồng: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			hopDongs = append(hopDongs, hd)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			http.Error(
				w,
				"Lỗi đọc danh sách hợp đồng: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		rows.Close()

		// =========================
		// 3. CHẤM CÔNG
		// =========================

		var chamCongs []ThongTinChamCong

		rows, err = db.Query(`
			SELECT ngay_cham_cong,
				COALESCE(gio_vao, ''),
				COALESCE(gio_ra, ''),
				COALESCE(trang_thai, ''),
				COALESCE(ghi_chu, '')
			FROM chamcong
			WHERE ma_nv = ?
			ORDER BY ngay_cham_cong DESC
		`, maNV)

		if err != nil {
			http.Error(
				w,
				"Lỗi truy vấn chấm công: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		for rows.Next() {
			var cc ThongTinChamCong

			err := rows.Scan(
				&cc.NgayChamCong,
				&cc.GioVao,
				&cc.GioRa,
				&cc.TrangThai,
				&cc.GhiChu,
			)

			if err != nil {
				rows.Close()
				http.Error(
					w,
					"Lỗi đọc dữ liệu chấm công: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			chamCongs = append(chamCongs, cc)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			http.Error(
				w,
				"Lỗi đọc danh sách chấm công: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		rows.Close()

		// =========================
		// 4. NGHỈ PHÉP
		// =========================

		var nghiPheps []ThongTinNghiPhep

		rows, err = db.Query(`
			SELECT ngay_bat_dau, ngay_ket_thuc,
			       loai_nghi, ly_do, trang_thai, ghi_chu
			FROM nghiphep
			WHERE ma_nv = ?
			ORDER BY ngay_bat_dau DESC
		`, maNV)

		if err != nil {
			http.Error(
				w,
				"Lỗi truy vấn nghỉ phép: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		for rows.Next() {
			var np ThongTinNghiPhep

			err := rows.Scan(
				&np.NgayBatDau,
				&np.NgayKetThuc,
				&np.LoaiNghi,
				&np.LyDo,
				&np.TrangThai,
				&np.GhiChu,
			)

			if err != nil {
				rows.Close()
				http.Error(
					w,
					"Lỗi đọc dữ liệu nghỉ phép: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			nghiPheps = append(nghiPheps, np)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			http.Error(
				w,
				"Lỗi đọc danh sách nghỉ phép: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		rows.Close()

		// =========================
		// 5. BẢNG LƯƠNG
		// =========================

		var bangLuongs []ThongTinBangLuong

		rows, err = db.Query(`
			SELECT thang, nam, luong_co_ban,
			       so_ngay_cong, phu_cap, khau_tru,
			       tong_luong, ghi_chu
			FROM bangluong
			WHERE ma_nv = ?
			ORDER BY nam DESC, thang DESC
		`, maNV)

		if err != nil {
			http.Error(
				w,
				"Lỗi truy vấn bảng lương: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		for rows.Next() {
			var bl ThongTinBangLuong

			err := rows.Scan(
				&bl.Thang,
				&bl.Nam,
				&bl.LuongCoBan,
				&bl.SoNgayCong,
				&bl.PhuCap,
				&bl.KhauTru,
				&bl.TongLuong,
				&bl.GhiChu,
			)

			if err != nil {
				rows.Close()
				http.Error(
					w,
					"Lỗi đọc dữ liệu bảng lương: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			bangLuongs = append(bangLuongs, bl)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			http.Error(
				w,
				"Lỗi đọc danh sách bảng lương: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		rows.Close()

		// =========================
		// GỬI TOÀN BỘ DỮ LIỆU RA HTML
		// =========================

		data := struct {
			DaTim      bool
			NhanVien   *models.NhanVien
			HopDongs   []ThongTinHopDong
			ChamCongs  []ThongTinChamCong
			NghiPheps  []ThongTinNghiPhep
			BangLuongs []ThongTinBangLuong
			Error      string
		}{
			DaTim:      true,
			NhanVien:   &nv,
			HopDongs:   hopDongs,
			ChamCongs:  chamCongs,
			NghiPheps:  nghiPheps,
			BangLuongs: bangLuongs,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(
				w,
				"Lỗi hiển thị thông tin nhân viên: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}
	}
}

// Hiển thị menu quản lý nhân viên
func QuanLyNhanVien() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(
			template.ParseFiles("templates/nhanvien/index.html"),
		)

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, "Lỗi hiển thị quản lý nhân viên", http.StatusInternalServerError)
		}
	}
}
