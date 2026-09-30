package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"qlnv/models"
)

// Danh sách chấm công
func DanhSachChamCong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, ma_nv, ngay_cham_cong,
			       gio_vao, gio_ra, trang_thai, ghi_chu
			FROM chamcong
			ORDER BY ngay_cham_cong DESC, id DESC
		`)

		if err != nil {
			http.Error(w, "Lỗi truy vấn database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var danhSach []models.ChamCong

		for rows.Next() {
			var cc models.ChamCong

			err := rows.Scan(
				&cc.ID,
				&cc.MaNV,
				&cc.NgayChamCong,
				&cc.GioVao,
				&cc.GioRa,
				&cc.TrangThai,
				&cc.GhiChu,
			)

			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(), http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, cc)
		}

		tmpl := template.Must(template.ParseFiles(
			"templates/chamcong/list.html",
		))

		tmpl.Execute(w, danhSach)
	}
}