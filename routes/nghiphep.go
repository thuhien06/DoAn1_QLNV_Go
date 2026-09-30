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
