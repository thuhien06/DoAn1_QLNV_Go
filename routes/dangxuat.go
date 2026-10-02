package routes

import "net/http"

func DangXuat(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "qlnv-session")
	if err != nil {
		http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
		return
	}

	// Xóa session
	session.Values["da_dang_nhap"] = false
	session.Values["ten_dang_nhap"] = nil
	session.Values["vai_tro"] = nil

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
}
