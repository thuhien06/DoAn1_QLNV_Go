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

// Thêm bảng lương
func ThemBangLuong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			tmpl := template.Must(
				template.ParseFiles("templates/bangluong/them.html"),
			)

			tmpl.Execute(w, nil)
			return
		}

		if r.Method == http.MethodPost {

			maNV := r.FormValue("ma_nv")
			thang := r.FormValue("thang")
			nam := r.FormValue("nam")
			luongCoBan := r.FormValue("luong_co_ban")
			soNgayCong := r.FormValue("so_ngay_cong")
			phuCap := r.FormValue("phu_cap")
			khauTru := r.FormValue("khau_tru")
			ghiChu := r.FormValue("ghi_chu")

			// Kiểm tra mã nhân viên
			var count int

			err := db.QueryRow(
				"SELECT COUNT(*) FROM nhanvien WHERE ma_nv = ?",
				maNV,
			).Scan(&count)

			if err != nil {
				http.Error(
					w,
					"Lỗi kiểm tra mã nhân viên",
					http.StatusInternalServerError,
				)
				return
			}

			if count == 0 {
				tmpl := template.Must(
					template.ParseFiles(
						"templates/bangluong/them.html",
					),
				)

				data := struct {
					Error string
				}{
					Error: "Mã nhân viên " + maNV + " không tồn tại!",
				}

				tmpl.Execute(w, data)
				return
			}

			// Tính tổng lương
			luong, _ := strconv.ParseFloat(luongCoBan, 64)
			phuCapValue, _ := strconv.ParseFloat(phuCap, 64)
			khauTruValue, _ := strconv.ParseFloat(khauTru, 64)

			tongLuong := luong + phuCapValue - khauTruValue

			_, err = db.Exec(`
				INSERT INTO bangluong
				(ma_nv, thang, nam, luong_co_ban,
				 so_ngay_cong, phu_cap, khau_tru,
				 tong_luong, ghi_chu)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				maNV,
				thang,
				nam,
				luongCoBan,
				soNgayCong,
				phuCap,
				khauTru,
				tongLuong,
				ghiChu,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/bangluong", http.StatusSeeOther)
		}
	}
}

// Sửa bảng lương
func SuaBangLuong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {

			var bl models.BangLuong

			err := db.QueryRow(`
				SELECT id, ma_nv, thang, nam, luong_co_ban,
				       so_ngay_cong, phu_cap, khau_tru,
				       tong_luong, ghi_chu
				FROM bangluong
				WHERE id = ?
			`, id).Scan(
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

			tmpl := template.Must(
				template.ParseFiles("templates/bangluong/sua.html"),
			)

			tmpl.Execute(w, bl)
			return
		}

		if r.Method == http.MethodPost {

			maNV := r.FormValue("ma_nv")
			thang := r.FormValue("thang")
			nam := r.FormValue("nam")
			luongCoBan := r.FormValue("luong_co_ban")
			soNgayCong := r.FormValue("so_ngay_cong")
			phuCap := r.FormValue("phu_cap")
			khauTru := r.FormValue("khau_tru")
			ghiChu := r.FormValue("ghi_chu")

			// Tính lại tổng lương
			luong, _ := strconv.ParseFloat(luongCoBan, 64)
			phuCapValue, _ := strconv.ParseFloat(phuCap, 64)
			khauTruValue, _ := strconv.ParseFloat(khauTru, 64)

			tongLuong := luong + phuCapValue - khauTruValue

			_, err := db.Exec(`
				UPDATE bangluong
				SET ma_nv = ?,
				    thang = ?,
				    nam = ?,
				    luong_co_ban = ?,
				    so_ngay_cong = ?,
				    phu_cap = ?,
				    khau_tru = ?,
				    tong_luong = ?,
				    ghi_chu = ?
				WHERE id = ?
			`,
				maNV,
				thang,
				nam,
				luongCoBan,
				soNgayCong,
				phuCap,
				khauTru,
				tongLuong,
				ghiChu,
				id,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/bangluong", http.StatusSeeOther)
		}
	}
}

// Xóa bảng lương
func XoaBangLuong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.Atoi(r.URL.Query().Get("id"))

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			"DELETE FROM bangluong WHERE id = ?",
			id,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/bangluong", http.StatusSeeOther)
	}
}
