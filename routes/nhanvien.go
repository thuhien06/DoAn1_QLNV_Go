package routes

import (
	"database/sql"
	"html/template"
	"net/http"

	"qlnv/models"
)

func DanhSachNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
			       so_dien_thoai, email, dia_chi, ngay_vao_lam,
			       phong_ban, chuc_vu
			FROM nhanvien
		`)

		if err != nil {
			http.Error(w, "Loi truy van database", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var danhSach []models.NhanVien

		for rows.Next() {
			var nv models.NhanVien

			err := rows.Scan(
				&nv.ID,
				&nv.MaNV,
				&nv.HoTen,
				&nv.NgaySinh,
				&nv.GioiTinh,
				&nv.SoDienThoai,
				&nv.Email,
				&nv.DiaChi,
				&nv.NgayVaoLam,
				&nv.PhongBan,
				&nv.ChucVu,
			)

			if err != nil {
				http.Error(w, "Loi doc du lieu", http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, nv)
		}

		tmpl := template.Must(template.ParseFiles(
			"templates/nhanvien/list.html",
		))

		tmpl.Execute(w, danhSach)
	}
}