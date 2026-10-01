package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"qlnv/models"
)

// Danh sách bảng lương
func DanhSachBangLuong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, ma_nv, thang, nam, luong_co_ban,
			       so_ngay_cong, phu_cap, khau_tru,
			       tong_luong, ghi_chu
			FROM bangluong
			ORDER BY nam DESC, thang DESC, id DESC
		`)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.BangLuong

		for rows.Next() {
			var bl models.BangLuong

			err := rows.Scan(
				&bl.ID,
				&bl.MaNV,
				&bl.Thang,
				&bl.Nam,
				&bl.LuongCoBan,
				&bl.SoNgayCong,
				&bl.PhuCap,
				&bl.KhauTru,
				&bl.TongLuong,
				&bl.GhiChu,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, bl)
		}

		tmpl := template.Must(
			template.ParseFiles("templates/bangluong/list.html"),
		)

		tmpl.Execute(w, danhSach)
	}
}
