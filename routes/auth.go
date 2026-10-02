package routes

import "net/http"

// Kiểm tra người dùng đã đăng nhập chưa
func DaDangNhap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		daDangNhap, ok := session.Values["da_dang_nhap"].(bool)

		if !ok || !daDangNhap {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

// Kiểm tra người dùng có đúng vai trò không
func CoVaiTro(vaiTroChoPhep ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			session, err := Store.Get(r, "qlnv-session")
			if err != nil {
				http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
				return
			}

			vaiTro, ok := session.Values["vai_tro"].(string)
			if !ok {
				http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
				return
			}

			// Kiểm tra role
			for _, role := range vaiTroChoPhep {
				if vaiTro == role {
					next(w, r)
					return
				}
			}

			// Không có quyền
			http.Error(w, "Bạn không có quyền truy cập chức năng này!", http.StatusForbidden)
		}
	}
}
