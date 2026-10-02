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