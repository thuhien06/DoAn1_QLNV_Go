package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"qlnv/models"
	"strings"
)

// Danh sách nghỉ phép
func DanhSachNghiPhep(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maNV := strings.TrimSpace(r.URL.Query().Get("ma_nv"))
		daTim := maNV != ""

		query := `
			SELECT id, ma_nv, ngay_bat_dau, ngay_ket_thuc,
			       loai_nghi, ly_do, trang_thai, ghi_chu
			FROM nghiphep
		`

		var args []interface{}

		if daTim {
			query += ` WHERE ma_nv = ?`
			args = append(args, maNV)
		} else {
			query += ` WHERE trang_thai = ?`
			args = append(args, "Chờ duyệt")
		}

		query += ` ORDER BY id DESC`

		rows, err := db.Query(query, args...)
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

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			DanhSach []models.NghiPhep
			MaNV     string
			DaTim    bool
			SoDon    int
		}{
			DanhSach: danhSach,
			MaNV:     maNV,
			DaTim:    daTim,
			SoDon:    len(danhSach),
		}

		tmpl := template.Must(
			template.ParseFiles("templates/nghiphep/list.html"),
		)

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Thêm nghỉ phép
func ThemNghiPhep(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		vaiTro, _ := session.Values["vai_tro"].(string)

		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		if r.Method == http.MethodGet {
			tmpl := template.Must(
				template.ParseFiles("templates/nghiphep/them.html"),
			)
			tmpl.Execute(w, nil)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Dữ liệu biểu mẫu không hợp lệ", http.StatusBadRequest)
			return
		}

		var maNV string

		if vaiTro == "nhanvien" {
			// Nhân viên chỉ được gửi đơn cho chính mình.
			maNV, _ = session.Values["ma_nv"].(string)
			if maNV == "" {
				http.Error(w, "Tài khoản chưa được liên kết với nhân viên", http.StatusForbidden)
				return
			}
		} else if vaiTro == "admin" || vaiTro == "hr" {
			// Admin/HR được tạo đơn cho nhân viên theo mã nhập vào.
			maNV = strings.TrimSpace(r.FormValue("ma_nv"))
		} else {
			http.Error(w, "Bạn không có quyền gửi đơn nghỉ phép", http.StatusForbidden)
			return
		}

		ngayBatDau := r.FormValue("ngay_bat_dau")
		ngayKetThuc := r.FormValue("ngay_ket_thuc")
		loaiNghi := r.FormValue("loai_nghi")
		lyDo := strings.TrimSpace(r.FormValue("ly_do"))
		ghiChu := strings.TrimSpace(r.FormValue("ghi_chu"))

		// Kiểm tra ngày hợp lệ và ngày kết thúc không trước ngày bắt đầu.
		batDau, err1 := time.Parse("2006-01-02", ngayBatDau)
		ketThuc, err2 := time.Parse("2006-01-02", ngayKetThuc)

		if err1 != nil || err2 != nil || ketThuc.Before(batDau) {
			tmpl := template.Must(
				template.ParseFiles("templates/nghiphep/them.html"),
			)
			tmpl.Execute(w, struct {
				Error string
			}{
				Error: "Ngày nghỉ không hợp lệ. Ngày kết thúc phải bằng hoặc sau ngày bắt đầu.",
			})
			return
		}

		var count int
		err = db.QueryRow(
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
			tmpl.Execute(w, struct {
				Error string
			}{
				Error: "Mã nhân viên không tồn tại!",
			})
			return
		}

		// Mọi đơn mới đều bắt đầu ở trạng thái Chờ duyệt.
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
			"Chờ duyệt",
			ghiChu,
		)

		if err != nil {
			http.Error(w, "Lỗi thêm đơn nghỉ phép: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if vaiTro == "nhanvien" {
			http.Redirect(w, r, "/nghiphep-ca-nhan", http.StatusSeeOther)
		} else {
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

func NghiPhepCaNhan(db *sql.DB) http.HandlerFunc {
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
			SELECT id, ma_nv, ngay_bat_dau, ngay_ket_thuc,
			       loai_nghi, ly_do, trang_thai, ghi_chu
			FROM nghiphep
			WHERE ma_nv = ?
			ORDER BY ngay_bat_dau DESC
		`, maNV)

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
			template.ParseFiles("templates/nghiphep/canhan.html"),
		)

		err = tmpl.Execute(w, danhSach)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
