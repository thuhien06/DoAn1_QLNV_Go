package routes

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qlnv/models"
)

// Danh sách chấm công theo mã nhân viên và tháng
func DanhSachChamCong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maNV := r.URL.Query().Get("ma_nv")
		thang := r.URL.Query().Get("thang")
		nam := r.URL.Query().Get("nam")

		// Chưa tìm kiếm thì chỉ hiện biểu mẫu
		tmpl := template.Must(template.ParseFiles(
			"templates/chamcong/list.html",
		))

		type DongChamCong struct {
			models.ChamCong
			DanhGia string
		}

		type DuLieuTrang struct {
			DanhSach []DongChamCong
			MaNV     string
			Thang    string
			Nam      string
			DaTim    bool
			Error    string
			SoCong   int
		}

		data := DuLieuTrang{
			MaNV:  maNV,
			Thang: thang,
			Nam:   nam,
		}

		if maNV == "" && thang == "" && nam == "" {
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		data.DaTim = true

		// Kiểm tra dữ liệu tìm kiếm
		thangSo, err := strconv.Atoi(thang)
		if err != nil || thangSo < 1 || thangSo > 12 {
			data.Error = "Tháng không hợp lệ."
			tmpl.Execute(w, data)
			return
		}

		namSo, err := strconv.Atoi(nam)
		if err != nil || namSo < 2000 || namSo > 2100 {
			data.Error = "Năm không hợp lệ."
			tmpl.Execute(w, data)
			return
		}

		// Kiểm tra mã nhân viên tồn tại
		var count int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM nhanvien WHERE ma_nv = ?",
			maNV,
		).Scan(&count)

		if err != nil {
			http.Error(w, "Lỗi kiểm tra nhân viên: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		if count == 0 {
			data.Error = "Mã nhân viên không tồn tại."
			tmpl.Execute(w, data)
			return
		}

		// Lọc dữ liệu theo tháng và năm
		tuNgay := time.Date(namSo, time.Month(thangSo), 1, 0, 0, 0, 0, time.Local)
		denNgay := tuNgay.AddDate(0, 1, 0)

		rows, err := db.Query(`
			SELECT id, ma_nv, ngay_cham_cong,
			       gio_vao, gio_ra, trang_thai, ghi_chu
			FROM chamcong
			WHERE ma_nv = ?
			  AND ngay_cham_cong >= ?
			  AND ngay_cham_cong < ?
			ORDER BY ngay_cham_cong DESC, id DESC
		`, maNV, tuNgay.Format("2006-01-02"),
			denNgay.Format("2006-01-02"))

		if err != nil {
			http.Error(w, "Lỗi truy vấn database: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer rows.Close()

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

			// Đánh giá giờ làm
			danhGia := "Chưa đủ dữ liệu"

			if cc.GioVao != "" && cc.GioRa != "" {
				gioVao, errVao := time.Parse("15:04:05", cc.GioVao)
				if errVao != nil {
					gioVao, errVao = time.Parse("15:04", cc.GioVao)
				}

				gioRa, errRa := time.Parse("15:04:05", cc.GioRa)
				if errRa != nil {
					gioRa, errRa = time.Parse("15:04", cc.GioRa)
				}

				if errVao == nil && errRa == nil {
					var ketQua []string

					mocVao := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
					mocRa := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)

					if gioVao.After(mocVao) {
						ketQua = append(ketQua, "Đi muộn")
					}

					if gioRa.Before(mocRa) {
						ketQua = append(ketQua, "Về sớm")
					}

					if len(ketQua) == 0 {
						danhGia = "Đủ giờ"
					} else {
						danhGia = strings.Join(ketQua, ", ")
					}
				}
			}

			data.DanhSach = append(data.DanhSach, DongChamCong{
				ChamCong: cc,
				DanhGia:  danhGia,
			})
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		// Đếm số ngày có chấm công trong tháng
		ngayDaDem := make(map[string]bool)

		for _, dong := range data.DanhSach {
			if dong.NgayChamCong != nil {
				ngay := dong.NgayChamCong.Format("2006-01-02")
				ngayDaDem[ngay] = true
			}
		}

		data.SoCong = len(ngayDaDem)

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

			var gioVao, gioRa, trangThai, ghiChu sql.NullString

			err := db.QueryRow(`
        SELECT id, ma_nv, ngay_cham_cong,
               gio_vao, gio_ra, trang_thai, ghi_chu
        FROM chamcong
        WHERE id = ?
    `, id).Scan(
				&cc.ID,
				&cc.MaNV,
				&cc.NgayChamCong,
				&gioVao,
				&gioRa,
				&trangThai,
				&ghiChu,
			)

			if err != nil {
				http.Error(
					w,
					"Không đọc được dữ liệu chấm công: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			cc.GioVao = gioVao.String
			cc.GioRa = gioRa.String
			cc.TrangThai = trangThai.String
			cc.GhiChu = ghiChu.String

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
			http.Error(w, "Lỗi session", http.StatusInternalServerError)
			return
		}

		maNV, ok := session.Values["ma_nv"].(string)
		if !ok || maNV == "" {
			http.Error(w, "Không xác định được mã nhân viên", http.StatusUnauthorized)
			return
		}

		// Lấy thời gian hiện tại
		now := time.Now()
		thang := int(now.Month())
		nam := now.Year()

		// Nếu nhân viên chọn tháng/năm khác thì dùng giá trị đó
		if r.URL.Query().Get("thang") != "" {
			fmt.Sscanf(r.URL.Query().Get("thang"), "%d", &thang)
		}

		if r.URL.Query().Get("nam") != "" {
			fmt.Sscanf(r.URL.Query().Get("nam"), "%d", &nam)
		}

		query := `
			SELECT id, ma_nv, ngay_cham_cong, gio_vao, gio_ra,
			       trang_thai, ghi_chu
			FROM chamcong
			WHERE ma_nv = ?
			  AND MONTH(ngay_cham_cong) = ?
			  AND YEAR(ngay_cham_cong) = ?
			ORDER BY ngay_cham_cong DESC
		`

		rows, err := db.Query(query, maNV, thang, nam)
		if err != nil {
			http.Error(w, "Lỗi lấy dữ liệu chấm công: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.ChamCong

		for rows.Next() {
			var cc models.ChamCong

			var gioVao sql.NullString
			var gioRa sql.NullString
			var trangThai sql.NullString
			var ghiChu sql.NullString

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
				http.Error(w, "Lỗi đọc dữ liệu chấm công: "+err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Đếm số ngày chấm công khác nhau
		ngayDaCham := make(map[string]bool)

		for _, cc := range danhSach {
			if cc.NgayChamCong != nil {
				ngay := cc.NgayChamCong.Format("2006-01-02")
				ngayDaCham[ngay] = true
			}
		}

		tongSo := len(ngayDaCham)

		data := struct {
			DanhSach []models.ChamCong
			Thang    int
			Nam      int
			TongSo   int
		}{
			DanhSach: danhSach,
			Thang:    thang,
			Nam:      nam,
			TongSo:   tongSo,
		}

		tmpl := template.Must(
			template.ParseFiles("templates/chamcong/canhan.html"),
		)

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị trang chấm công cá nhân", http.StatusInternalServerError)
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
