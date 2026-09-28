package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"qlnv/models"
)

// Hiển thị danh sách nhân viên
func DanhSachNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Lấy từ khóa tìm kiếm
		maNV := r.URL.Query().Get("ma_nv")

		var rows *sql.Rows
		var err error

		// Nếu có mã nhân viên thì tìm kiếm
		if maNV != "" {
			rows, err = db.Query(`
				SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
				       so_dien_thoai, email, dia_chi, ngay_vao_lam,
				       phong_ban, chuc_vu
				FROM nhanvien
				WHERE ma_nv LIKE ?
			`, "%"+maNV+"%")
		} else {
			// Nếu không có từ khóa thì lấy toàn bộ nhân viên
			rows, err = db.Query(`
				SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
				       so_dien_thoai, email, dia_chi, ngay_vao_lam,
				       phong_ban, chuc_vu
				FROM nhanvien
			`)
		}

		// Kiểm tra lỗi truy vấn
		if err != nil {
			http.Error(w, "Lỗi truy vấn database", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		// Tạo danh sách nhân viên
		var danhSach []models.NhanVien

		// Duyệt dữ liệu
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

		// Đọc giao diện
		tmpl := template.Must(template.ParseFiles(
			"templates/nhanvien/list.html",
		))

		// Hiển thị danh sách
		tmpl.Execute(w, danhSach)
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
