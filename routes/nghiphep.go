package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

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

			// Kiểm tra mã nhân viên có tồn tại không
			var count int

			err := db.QueryRow(
				"SELECT COUNT(*) FROM nhanvien WHERE ma_nv = ?",
				maNV,
			).Scan(&count)

			if err != nil {
				http.Error(w, "Lỗi kiểm tra mã nhân viên", http.StatusInternalServerError)
				return
			}

			if count == 0 {
				tmpl := template.Must(
					template.ParseFiles("templates/nghiphep/them.html"),
				)

				data := struct {
					Error string
				}{
					Error: "Mã nhân viên " + maNV + " không tồn tại!",
				}

				tmpl.Execute(w, data)
				return
			}

			// Nếu mã nhân viên tồn tại thì mới thêm nghỉ phép
			_, err = db.Exec(`
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

// Sửa nghỉ phép
func SuaNghiPhep(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		if r.Method == "GET" {

			var np models.NghiPhep

			err := db.QueryRow(`
				SELECT id, ma_nv, ngay_bat_dau, ngay_ket_thuc,
				       loai_nghi, ly_do, trang_thai, ghi_chu
				FROM nghiphep
				WHERE id = ?
			`, id).Scan(
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

			tmpl := template.Must(
				template.ParseFiles("templates/nghiphep/sua.html"),
			)

			tmpl.Execute(w, np)
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
				UPDATE nghiphep
				SET ma_nv = ?,
				    ngay_bat_dau = ?,
				    ngay_ket_thuc = ?,
				    loai_nghi = ?,
				    ly_do = ?,
				    trang_thai = ?,
				    ghi_chu = ?
				WHERE id = ?
			`,
				maNV,
				ngayBatDau,
				ngayKetThuc,
				loaiNghi,
				lyDo,
				trangThai,
				ghiChu,
				id,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/nghiphep", http.StatusSeeOther)
		}
	}
}

// Xóa nghỉ phép
func XoaNghiPhep(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			"DELETE FROM nghiphep WHERE id = ?",
			id,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/nghiphep", http.StatusSeeOther)
	}
}
