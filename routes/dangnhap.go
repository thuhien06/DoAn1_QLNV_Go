package routes

import (
	"database/sql"
	"html/template"
	"net/http"
)

func DangNhap(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Hiển thị trang đăng nhập
		if r.Method == http.MethodGet {
			tmpl := template.Must(
				template.ParseFiles("templates/dangnhap.html"),
			)

			tmpl.Execute(w, nil)
			return
		}

		// Xử lý đăng nhập
		if r.Method == http.MethodPost {

			tenDangNhap := r.FormValue("ten_dang_nhap")
			matKhau := r.FormValue("mat_khau")

			var vaiTro string

			err := db.QueryRow(`
				SELECT vai_tro
				FROM taikhoan
				WHERE ten_dang_nhap = ?
				  AND mat_khau = ?
			`, tenDangNhap, matKhau).Scan(&vaiTro)

			// Sai tài khoản hoặc mật khẩu
			if err != nil {
				tmpl := template.Must(
					template.ParseFiles("templates/dangnhap.html"),
				)

				data := struct {
					Error string
				}{
					Error: "Tên đăng nhập hoặc mật khẩu không đúng!",
				}

				tmpl.Execute(w, data)
				return
			}

			var maNV string

			if vaiTro == "nhanvien" {
				err = db.QueryRow(`
					SELECT ma_nv
					FROM taikhoan
					WHERE ten_dang_nhap = ?
				`, tenDangNhap).Scan(&maNV)

				if err != nil {
					http.Error(w, "Tài khoản nhân viên chưa được liên kết", http.StatusInternalServerError)
					return
				}
			}

			// Tạo session
			session, err := Store.Get(r, "qlnv-session")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			session.Values["ten_dang_nhap"] = tenDangNhap
			session.Values["vai_tro"] = vaiTro
			session.Values["ma_nv"] = maNV
			session.Values["da_dang_nhap"] = true

			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Đăng nhập thành công
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}
