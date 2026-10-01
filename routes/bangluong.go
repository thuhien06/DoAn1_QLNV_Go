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
			soNgayCong := r.FormValue("so_ngay_cong")
			phuCap := r.FormValue("phu_cap")
			khauTru := r.FormValue("khau_tru")
			ghiChu := r.FormValue("ghi_chu")

			// Lấy lương cơ bản từ hợp đồng
			var luongCoBan float64

			err := db.QueryRow(`
				SELECT luong_co_ban
				FROM hopdong
				WHERE ma_nv = ?
				ORDER BY id DESC
				LIMIT 1
			`, maNV).Scan(&luongCoBan)

			if err != nil {
				tmpl := template.Must(
					template.ParseFiles(
						"templates/bangluong/them.html",
					),
				)

				data := struct {
					Error string
				}{
					Error: "Không tìm thấy hợp đồng hoặc lương cơ bản của nhân viên " + maNV,
				}

				tmpl.Execute(w, data)
				return
			}

			// Chuyển dữ liệu sang số
			thangValue, err := strconv.Atoi(thang)
			if err != nil {
				http.Error(w, "Tháng không hợp lệ", http.StatusBadRequest)
				return
			}

			namValue, err := strconv.Atoi(nam)
			if err != nil {
				http.Error(w, "Năm không hợp lệ", http.StatusBadRequest)
				return
			}

			soNgayCongValue, err := strconv.Atoi(soNgayCong)
			if err != nil {
				http.Error(w, "Số ngày công không hợp lệ", http.StatusBadRequest)
				return
			}

			phuCapValue, err := strconv.ParseFloat(phuCap, 64)
			if err != nil {
				phuCapValue = 0
			}

			khauTruValue, err := strconv.ParseFloat(khauTru, 64)
			if err != nil {
				khauTruValue = 0
			}

			// Tính lương theo ngày công
			luongTheoNgay := luongCoBan / 26

			tongLuong := luongTheoNgay*float64(soNgayCongValue) +
				phuCapValue -
				khauTruValue

			_, err = db.Exec(`
				INSERT INTO bangluong
				(ma_nv, thang, nam, luong_co_ban,
				 so_ngay_cong, phu_cap, khau_tru,
				 tong_luong, ghi_chu)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				maNV,
				thangValue,
				namValue,
				luongCoBan,
				soNgayCongValue,
				phuCapValue,
				khauTruValue,
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

		if r.Method == http.MethodPost {

			maNV := r.FormValue("ma_nv")
			thang := r.FormValue("thang")
			nam := r.FormValue("nam")
			soNgayCong := r.FormValue("so_ngay_cong")
			phuCap := r.FormValue("phu_cap")
			khauTru := r.FormValue("khau_tru")
			ghiChu := r.FormValue("ghi_chu")

			// Lấy lương cơ bản từ hợp đồng
			var luongCoBan float64

			err := db.QueryRow(`
				SELECT luong_co_ban
				FROM hopdong
				WHERE ma_nv = ?
				ORDER BY id DESC
				LIMIT 1
			`, maNV).Scan(&luongCoBan)

			if err != nil {
				http.Error(
					w,
					"Không tìm thấy hợp đồng hoặc lương cơ bản của nhân viên",
					http.StatusBadRequest,
				)
				return
			}

			thangValue, err := strconv.Atoi(thang)
			if err != nil {
				http.Error(w, "Tháng không hợp lệ", http.StatusBadRequest)
				return
			}

			namValue, err := strconv.Atoi(nam)
			if err != nil {
				http.Error(w, "Năm không hợp lệ", http.StatusBadRequest)
				return
			}

			soNgayCongValue, err := strconv.Atoi(soNgayCong)
			if err != nil {
				http.Error(w, "Số ngày công không hợp lệ", http.StatusBadRequest)
				return
			}

			phuCapValue, err := strconv.ParseFloat(phuCap, 64)
			if err != nil {
				phuCapValue = 0
			}

			khauTruValue, err := strconv.ParseFloat(khauTru, 64)
			if err != nil {
				khauTruValue = 0
			}

			// Tính lại lương
			luongTheoNgay := luongCoBan / 26

			tongLuong := luongTheoNgay*float64(soNgayCongValue) +
				phuCapValue -
				khauTruValue

			_, err = db.Exec(`
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
				thangValue,
				namValue,
				luongCoBan,
				soNgayCongValue,
				phuCapValue,
				khauTruValue,
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
