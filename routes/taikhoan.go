package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"qlnv/models"
)

// Danh sách tài khoản
func DanhSachTaiKhoan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, ten_dang_nhap, mat_khau, vai_tro, ma_nv
			FROM taikhoan
			ORDER BY id
		`)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var danhSach []models.TaiKhoan

		for rows.Next() {
			var tk models.TaiKhoan

			var maNV sql.NullString

			err := rows.Scan(
				&tk.ID,
				&tk.TenDangNhap,
				&tk.MatKhau,
				&tk.VaiTro,
				&maNV,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if maNV.Valid {
				tk.MaNV = maNV.String
			} else {
				tk.MaNV = ""
			}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, tk)
		}

		tmpl := template.Must(
			template.ParseFiles("templates/taikhoan/list.html"),
		)

		err = tmpl.Execute(w, danhSach)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Thêm tài khoản
func ThemTaiKhoan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {

			tmpl := template.Must(
				template.ParseFiles("templates/taikhoan/them.html"),
			)

			tmpl.Execute(w, nil)
			return
		}

		if r.Method == http.MethodPost {

			tenDangNhap := r.FormValue("ten_dang_nhap")
			matKhau := r.FormValue("mat_khau")
			vaiTro := r.FormValue("vai_tro")
			maNV := r.FormValue("ma_nv")

			// Nếu không phải nhân viên thì không cần mã nhân viên
			if vaiTro != "nhanvien" {
				maNV = ""
			}

			// Kiểm tra tên đăng nhập đã tồn tại
			var count int

			err := db.QueryRow(`
				SELECT COUNT(*)
				FROM taikhoan
				WHERE ten_dang_nhap = ?
			`, tenDangNhap).Scan(&count)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if count > 0 {
				http.Error(
					w,
					"Tên đăng nhập đã tồn tại!",
					http.StatusBadRequest,
				)
				return
			}

			// Nếu là nhân viên thì kiểm tra mã nhân viên
			if vaiTro == "nhanvien" {

				var employeeCount int

				err := db.QueryRow(`
					SELECT COUNT(*)
					FROM nhanvien
					WHERE ma_nv = ?
				`, maNV).Scan(&employeeCount)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if employeeCount == 0 {
					http.Error(
						w,
						"Mã nhân viên không tồn tại!",
						http.StatusBadRequest,
					)
					return
				}

				// Kiểm tra nhân viên đã có tài khoản chưa
				var accountCount int

				err = db.QueryRow(`
					SELECT COUNT(*)
					FROM taikhoan
					WHERE ma_nv = ?
				`, maNV).Scan(&accountCount)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if accountCount > 0 {
					http.Error(
						w,
						"Nhân viên này đã có tài khoản!",
						http.StatusBadRequest,
					)
					return
				}
			}

			_, err = db.Exec(`
				INSERT INTO taikhoan
				(ten_dang_nhap, mat_khau, vai_tro, ma_nv)
				VALUES (?, ?, ?, ?)
			`,
				tenDangNhap,
				matKhau,
				vaiTro,
				maNV,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(
				w,
				r,
				"/taikhoan",
				http.StatusSeeOther,
			)
		}
	}
}

// Sửa tài khoản
func SuaTaiKhoan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {

			var tk models.TaiKhoan

			err := db.QueryRow(`
				SELECT id, ten_dang_nhap, mat_khau, vai_tro, ma_nv
				FROM taikhoan
				WHERE id = ?
			`, id).Scan(
				&tk.ID,
				&tk.TenDangNhap,
				&tk.MatKhau,
				&tk.VaiTro,
				&tk.MaNV,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl := template.Must(
				template.ParseFiles("templates/taikhoan/sua.html"),
			)

			tmpl.Execute(w, tk)
			return
		}

		if r.Method == http.MethodPost {

			tenDangNhap := r.FormValue("ten_dang_nhap")
			matKhau := r.FormValue("mat_khau")
			vaiTro := r.FormValue("vai_tro")
			maNV := r.FormValue("ma_nv")

			if vaiTro != "nhanvien" {
				maNV = ""
			}

			_, err := db.Exec(`
				UPDATE taikhoan
				SET ten_dang_nhap = ?,
					mat_khau = ?,
					vai_tro = ?,
					ma_nv = ?
				WHERE id = ?
			`,
				tenDangNhap,
				matKhau,
				vaiTro,
				maNV,
				id,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(
				w,
				r,
				"/taikhoan",
				http.StatusSeeOther,
			)
		}
	}
}

// Xóa tài khoản
func XoaTaiKhoan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`
			DELETE FROM taikhoan
			WHERE id = ?
		`, id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(
			w,
			r,
			"/taikhoan",
			http.StatusSeeOther,
		)
	}
}
