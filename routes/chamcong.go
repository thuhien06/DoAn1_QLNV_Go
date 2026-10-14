package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

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
			var gioVao, gioRa, trangThai, ghiChu sql.NullString

			err := rows.Scan(
				&cc.ID,
				&cc.MaNV,
				&cc.NgayChamCong,
				&gioVao,
				&gioRa,
				&trangThai,
				&ghiChu,
			)
			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			if gioVao.Valid {
				cc.GioVao = gioVao.String
			}
			if gioRa.Valid {
				cc.GioRa = gioRa.String
			}
			if trangThai.Valid {
				cc.TrangThai = trangThai.String
			}
			if ghiChu.Valid {
				cc.GhiChu = ghiChu.String
			}

			danhSach = append(danhSach, cc)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles(
			"templates/chamcong/list.html",
		))

		tmpl.Execute(w, danhSach)
	}
}

// Thêm chấm công
func ThemChamCong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			tmpl := template.Must(template.ParseFiles(
				"templates/chamcong/them.html",
			))

			tmpl.Execute(w, nil)
			return
		}

		if r.Method == "POST" {
			maNV := r.FormValue("ma_nv")
			ngayChamCong := r.FormValue("ngay_cham_cong")
			gioVao := r.FormValue("gio_vao")
			gioRa := r.FormValue("gio_ra")
			trangThai := r.FormValue("trang_thai")
			ghiChu := r.FormValue("ghi_chu")

			//ktra ma nv co ton tai ko
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
					template.ParseFiles("templates/chamcong/them.html"),
				)

				data := struct {
					Error string
				}{
					Error: "Mã nhân viên " + maNV + " không tồn tại!",
				}

				tmpl.Execute(w, data)
				return
			}

			// Nếu mã nhân viên tồn tại thì mới thêm cham cong
			_, err = db.Exec(
				`INSERT INTO chamcong
				(ma_nv, ngay_cham_cong, gio_vao, gio_ra, trang_thai, ghi_chu)
				VALUES (?, ?, ?, ?, ?, ?)
				`,
				maNV,
				ngayChamCong,
				gioVao,
				gioRa,
				trangThai,
				ghiChu,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/chamcong", http.StatusSeeOther)
		}

	}
}

// Sửa chấm công
func SuaChamCong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.URL.Query().Get("id")

		if r.Method == http.MethodGet {

			var cc models.ChamCong

			err := db.QueryRow(`
				SELECT id, ma_nv, ngay_cham_cong,
				       gio_vao, gio_ra, trang_thai, ghi_chu
				FROM chamcong
				WHERE id = ?
			`, id).Scan(
				&cc.ID,
				&cc.MaNV,
				&cc.NgayChamCong,
				&cc.GioVao,
				&cc.GioRa,
				&cc.TrangThai,
				&cc.GhiChu,
			)

			if err != nil {
				http.Error(w, "Không tìm thấy dữ liệu chấm công", http.StatusNotFound)
				return
			}

			tmpl := template.Must(template.ParseFiles(
				"templates/chamcong/sua.html",
			))

			tmpl.Execute(w, cc)
			return
		}

		maNV := r.FormValue("ma_nv")
		ngayChamCong := r.FormValue("ngay_cham_cong")
		gioVao := r.FormValue("gio_vao")
		gioRa := r.FormValue("gio_ra")
		trangThai := r.FormValue("trang_thai")
		ghiChu := r.FormValue("ghi_chu")

		_, err := db.Exec(`
			UPDATE chamcong
			SET ma_nv = ?,
			    ngay_cham_cong = ?,
			    gio_vao = ?,
			    gio_ra = ?,
			    trang_thai = ?,
			    ghi_chu = ?
			WHERE id = ?
		`,
			maNV,
			ngayChamCong,
			gioVao,
			gioRa,
			trangThai,
			ghiChu,
			id,
		)

		if err != nil {
			http.Error(w, "Lỗi sửa chấm công: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/chamcong", http.StatusSeeOther)
	}
}

// Xóa chấm công
func XoaChamCong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.URL.Query().Get("id")

		idCC, err := strconv.Atoi(id)

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			"DELETE FROM chamcong WHERE id = ?",
			idCC,
		)

		if err != nil {
			http.Error(w, "Lỗi xóa chấm công: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/chamcong", http.StatusSeeOther)
	}
}

// chấm công cá nhân
func ChamCongCaNhan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		maNV, ok := session.Values["ma_nv"].(string)
		if !ok || maNV == "" {
			http.Error(w, "Tài khoản chưa được liên kết với nhân viên", http.StatusForbidden)
			return
		}

		rows, err := db.Query(`
			SELECT id, ma_nv, ngay_cham_cong,
			       gio_vao, gio_ra, trang_thai, ghi_chu
			FROM chamcong
			WHERE ma_nv = ?
			ORDER BY ngay_cham_cong DESC
		`, maNV)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.ChamCong

		for rows.Next() {
			var cc models.ChamCong
			var gioVao, gioRa, trangThai, ghiChu sql.NullString

			err := rows.Scan(
				&cc.ID,
				&cc.MaNV,
				&cc.NgayChamCong,
				&gioVao,
				&gioRa,
				&trangThai,
				&ghiChu,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if gioVao.Valid {
				cc.GioVao = gioVao.String
			}
			if gioRa.Valid {
				cc.GioRa = gioRa.String
			}
			if trangThai.Valid {
				cc.TrangThai = trangThai.String
			}
			if ghiChu.Valid {
				cc.GhiChu = ghiChu.String
			}

			danhSach = append(danhSach, cc)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(
			template.ParseFiles("templates/chamcong/canhan.html"),
		)

		if err = tmpl.Execute(w, danhSach); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Nhân viên chấm công vào
func ChamCongVao(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không hợp lệ", http.StatusMethodNotAllowed)
			return
		}

		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		maNV, ok := session.Values["ma_nv"].(string)
		if !ok || maNV == "" {
			http.Error(w, "Tài khoản chưa được liên kết với nhân viên", http.StatusForbidden)
			return
		}

		ngay := time.Now().Format("2006-01-02")
		gio := time.Now().Format("15:04:05")

		var id int
		err = db.QueryRow(`
			SELECT id FROM chamcong
			WHERE ma_nv = ? AND ngay_cham_cong = ?
		`, maNV, ngay).Scan(&id)

		if err == nil {
			http.Error(w, "Hôm nay bạn đã có bản ghi chấm công", http.StatusConflict)
			return
		}
		if err != sql.ErrNoRows {
			http.Error(w, "Lỗi kiểm tra chấm công", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			INSERT INTO chamcong
				(ma_nv, ngay_cham_cong, gio_vao, trang_thai)
			VALUES (?, ?, ?, ?)
		`, maNV, ngay, gio, "Đã chấm công vào")

		if err != nil {
			http.Error(w, "Lỗi chấm công vào: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/chamcong-ca-nhan", http.StatusSeeOther)
	}
}

// Nhân viên chấm công ra
func ChamCongRa(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không hợp lệ", http.StatusMethodNotAllowed)
			return
		}

		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		maNV, ok := session.Values["ma_nv"].(string)
		if !ok || maNV == "" {
			http.Error(w, "Tài khoản chưa được liên kết với nhân viên", http.StatusForbidden)
			return
		}

		ngay := time.Now().Format("2006-01-02")
		gio := time.Now().Format("15:04:05")

		result, err := db.Exec(`
			UPDATE chamcong
			SET gio_ra = ?, trang_thai = 'Đã chấm công ra'
			WHERE ma_nv = ?
			  AND ngay_cham_cong = ?
			  AND gio_vao IS NOT NULL
			  AND gio_ra IS NULL
		`, gio, maNV, ngay)

		if err != nil {
			http.Error(w, "Lỗi chấm công ra: "+err.Error(), http.StatusInternalServerError)
			return
		}

		n, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Không kiểm tra được kết quả chấm công", http.StatusInternalServerError)
			return
		}
		if n == 0 {
			http.Error(w, "Bạn chưa chấm công vào hoặc đã chấm công ra hôm nay", http.StatusConflict)
			return
		}

		http.Redirect(w, r, "/chamcong-ca-nhan", http.StatusSeeOther)
	}
}
