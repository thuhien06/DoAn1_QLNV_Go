package routes

import (
	"database/sql"
	"html/template"
	"net/http"

	"qlnv/models"
)

// Danh sách nghỉ phép
func DanhSachNghiPhep(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, ma_nv, ngay_bat_dau, ngay_ket_thuc,
			       loai_nghi, ly_do, trang_thai, ghi_chu
			FROM nghiphep
			ORDER BY id DESC
		`)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.NghiPhep

		for rows.Next() {
			var np models.NghiPhep

			err := rows.Scan(
				&np.ID,
				&np.MaNV,
				&np.NgayBatDau,
				&np.NgayKetThuc,
				&np.LoaiNghi,
				&np.LyDo,
				&np.TrangThai,
				&np.GhiChu,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, np)
		}

		tmpl := template.Must(
			template.ParseFiles("templates/nghiphep/list.html"),
		)

		tmpl.Execute(w, danhSach)
	}
}

// Thêm nghỉ phép
func ThemNghiPhep(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			tmpl := template.Must(
				template.ParseFiles("templates/nghiphep/them.html"),
			)

			tmpl.Execute(w, nil)
			return
		}

		if r.Method == "POST" {

			maNV := r.FormValue("ma_nv")
			ngayBatDau := r.FormValue("ngay_bat_dau")
			ngayKetThuc := r.FormValue("ngay_ket_thuc")
			loaiNghi := r.FormValue("loai_nghi")
			lyDo := r.FormValue("ly_do")
			trangThai := r.FormValue("trang_thai")
			ghiChu := r.FormValue("ghi_chu")

			_, err := db.Exec(`
				INSERT INTO nghiphep
				(ma_nv, ngay_bat_dau, ngay_ket_thuc,
				 loai_nghi, ly_do, trang_thai, ghi_chu)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`,
				maNV,
				ngayBatDau,
				ngayKetThuc,
				loaiNghi,
				lyDo,
				trangThai,
				ghiChu,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/nghiphep", http.StatusSeeOther)
		}
	}
}

