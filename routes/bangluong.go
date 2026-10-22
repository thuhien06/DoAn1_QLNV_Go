package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

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

// Thêm bảng lương - bước xem trước
func ThemBangLuong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {

			now := time.Now()

			data := struct {
				Error string
				Thang int
				Nam   int
			}{
				Error: "",
				Thang: int(now.Month()),
				Nam:   now.Year(),
			}

			tmpl := template.Must(
				template.ParseFiles("templates/bangluong/them.html"),
			)

			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Lỗi hiển thị: "+err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		maNV := r.FormValue("ma_nv")
		thang := r.FormValue("thang")
		nam := r.FormValue("nam")
		phuCap := r.FormValue("phu_cap")
		khauTru := r.FormValue("khau_tru")
		ghiChu := r.FormValue("ghi_chu")

		// Kiểm tra tháng
		thangValue, err := strconv.Atoi(thang)
		if err != nil || thangValue < 1 || thangValue > 12 {
			http.Error(w, "Tháng không hợp lệ", http.StatusBadRequest)
			return
		}

		// Kiểm tra năm
		namValue, err := strconv.Atoi(nam)
		if err != nil || namValue < 2000 {
			http.Error(w, "Năm không hợp lệ", http.StatusBadRequest)
			return
		}

		// Kiểm tra nhân viên
		var tonTaiNhanVien int

		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM nhanvien
			WHERE ma_nv = ?
		`, maNV).Scan(&tonTaiNhanVien)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if tonTaiNhanVien == 0 {
			tmpl := template.Must(
				template.ParseFiles("templates/bangluong/them.html"),
			)

			tmpl.Execute(w, struct {
				Error string
			}{
				Error: "Không tìm thấy nhân viên có mã " + maNV,
			})
			return
		}

		// Kiểm tra bảng lương đã tồn tại chưa
		var daTonTai int

		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM bangluong
			WHERE ma_nv = ?
			  AND thang = ?
			  AND nam = ?
		`, maNV, thangValue, namValue).Scan(&daTonTai)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if daTonTai > 0 {
			tmpl := template.Must(
				template.ParseFiles("templates/bangluong/them.html"),
			)

			tmpl.Execute(w, struct {
				Error string
			}{
				Error: "Nhân viên " + maNV +
					" đã có bảng lương tháng " +
					strconv.Itoa(thangValue) + "/" +
					strconv.Itoa(namValue),
			})
			return
		}

		// =========================================================
		// TÍNH SỐ NGÀY CÔNG CHUẨN TRONG THÁNG
		// Chỉ tính từ thứ 2 đến thứ 6.
		// =========================================================

		ngayDauThang := time.Date(
			namValue,
			time.Month(thangValue),
			1,
			0, 0, 0, 0,
			time.Local,
		)

		ngayCuoiThang := ngayDauThang.AddDate(0, 1, -1)

		soNgayCongChuan := 0

		for ngay := ngayDauThang; !ngay.After(ngayCuoiThang); ngay = ngay.AddDate(0, 0, 1) {
			thu := ngay.Weekday()

			if thu == time.Saturday || thu == time.Sunday {
				continue
			}

			soNgayCongChuan++
		}

		// =========================================================
		// LẤY NGÀY CHẤM CÔNG THỰC TẾ
		// Chỉ tính thứ 2 đến thứ 6.
		// =========================================================

		rowsCong, err := db.Query(`
			SELECT DISTINCT ngay_cham_cong
			FROM chamcong
			WHERE ma_nv = ?
			  AND MONTH(ngay_cham_cong) = ?
			  AND YEAR(ngay_cham_cong) = ?
			  AND ngay_cham_cong IS NOT NULL
		`, maNV, thangValue, namValue)

		if err != nil {
			http.Error(w, "Lỗi đọc dữ liệu chấm công: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer rowsCong.Close()

		ngayCongDaTinh := make(map[string]bool)

		for rowsCong.Next() {
			var ngay time.Time

			if err := rowsCong.Scan(&ngay); err != nil {
				http.Error(w, "Lỗi đọc dữ liệu chấm công: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			thu := ngay.Weekday()

			if thu == time.Saturday || thu == time.Sunday {
				continue
			}

			ngayCongDaTinh[ngay.Format("2006-01-02")] = true
		}

		if err := rowsCong.Err(); err != nil {
			http.Error(w, "Lỗi đọc dữ liệu chấm công: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		// =========================================================
		// LẤY NGÀY NGHỈ PHÉP ĐÃ DUYỆT
		// Chỉ tính thứ 2 đến thứ 6.
		// =========================================================

		rowsPhep, err := db.Query(`
			SELECT ngay_bat_dau, ngay_ket_thuc
			FROM nghiphep
			WHERE ma_nv = ?
			  AND trang_thai = 'Đã duyệt'
			  AND ngay_bat_dau <= ?
			  AND ngay_ket_thuc >= ?
		`,
			maNV,
			ngayCuoiThang,
			ngayDauThang,
		)

		if err != nil {
			http.Error(w, "Lỗi đọc dữ liệu nghỉ phép: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer rowsPhep.Close()

		ngayPhepDaTinh := make(map[string]bool)

		for rowsPhep.Next() {
			var batDau time.Time
			var ketThuc time.Time

			if err := rowsPhep.Scan(&batDau, &ketThuc); err != nil {
				http.Error(w, "Lỗi đọc dữ liệu nghỉ phép: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			// Chỉ lấy phần nghỉ nằm trong tháng đang tính lương.
			if batDau.Before(ngayDauThang) {
				batDau = ngayDauThang
			}

			if ketThuc.After(ngayCuoiThang) {
				ketThuc = ngayCuoiThang
			}

			for ngay := batDau; !ngay.After(ketThuc); ngay = ngay.AddDate(0, 0, 1) {
				thu := ngay.Weekday()

				if thu == time.Saturday || thu == time.Sunday {
					continue
				}

				ngayPhepDaTinh[ngay.Format("2006-01-02")] = true
			}
		}

		if err := rowsPhep.Err(); err != nil {
			http.Error(w, "Lỗi đọc dữ liệu nghỉ phép: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		// =========================================================
		// TỔNG NGÀY CÔNG HƯỞNG LƯƠNG
		// Chấm công + nghỉ phép đã duyệt.
		// =========================================================

		ngayHuongLuong := make(map[string]bool)

		for ngay := range ngayCongDaTinh {
			ngayHuongLuong[ngay] = true
		}

		for ngay := range ngayPhepDaTinh {
			ngayHuongLuong[ngay] = true
		}

		soNgayCong := len(ngayHuongLuong)

		// Không cho số công hưởng lương vượt quá số công chuẩn.
		if soNgayCong > soNgayCongChuan {
			soNgayCong = soNgayCongChuan
		}

		// =========================================================
		// LẤY LƯƠNG CƠ BẢN TỪ HỢP ĐỒNG
		// =========================================================

		var luongCoBan float64

		err = db.QueryRow(`
			SELECT luong_co_ban
			FROM hopdong
			WHERE ma_nv = ?
			  AND nhom_hop_dong = 'lao_dong'
			ORDER BY id DESC
			LIMIT 1
		`, maNV).Scan(&luongCoBan)

		if err != nil {
			tmpl := template.Must(
				template.ParseFiles("templates/bangluong/them.html"),
			)

			tmpl.Execute(w, struct {
				Error string
			}{
				Error: "Không tìm thấy hợp đồng nhân sự có lương cơ bản cho nhân viên " + maNV,
			})
			return
		}

		// =========================================================
		// PHỤ CẤP
		// =========================================================

		phuCapValue := 0.0

		if phuCap != "" {
			phuCapValue, err = strconv.ParseFloat(phuCap, 64)

			if err != nil || phuCapValue < 0 {
				http.Error(w, "Phụ cấp không hợp lệ", http.StatusBadRequest)
				return
			}
		}

		// =========================================================
		// KHẤU TRỪ
		// =========================================================

		khauTruValue := 0.0

		if khauTru != "" {
			khauTruValue, err = strconv.ParseFloat(khauTru, 64)

			if err != nil || khauTruValue < 0 {
				http.Error(w, "Khấu trừ không hợp lệ", http.StatusBadRequest)
				return
			}
		}

		// =========================================================
		// TÍNH TỔNG LƯƠNG
		// =========================================================

		luongTheoNgay := 0.0

		if soNgayCongChuan > 0 {
			luongTheoNgay = luongCoBan / float64(soNgayCongChuan)
		}

		tongLuong := luongTheoNgay*float64(soNgayCong) +
			phuCapValue -
			khauTruValue

		// =========================================================
		// DỮ LIỆU CHUYỂN SANG TRANG XEM TRƯỚC
		// =========================================================

		data := models.BangLuong{
			MaNV:       maNV,
			Thang:      thangValue,
			Nam:        namValue,
			LuongCoBan: luongCoBan,
			SoNgayCong: soNgayCong,
			PhuCap:     phuCapValue,
			KhauTru:    khauTruValue,
			TongLuong:  tongLuong,
			GhiChu:     ghiChu,
		}

		tmpl := template.Must(
			template.ParseFiles("templates/bangluong/xemtruoc.html"),
		)

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị xem trước: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}
}

// Lưu bảng lương sau khi HR xác nhận xem trước
func LuuBangLuong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		maNV := r.FormValue("ma_nv")

		thang, err := strconv.Atoi(r.FormValue("thang"))
		if err != nil || thang < 1 || thang > 12 {
			http.Error(w, "Tháng không hợp lệ", http.StatusBadRequest)
			return
		}

		nam, err := strconv.Atoi(r.FormValue("nam"))
		if err != nil || nam < 2000 {
			http.Error(w, "Năm không hợp lệ", http.StatusBadRequest)
			return
		}

		luongCoBan, err := strconv.ParseFloat(r.FormValue("luong_co_ban"), 64)
		if err != nil {
			http.Error(w, "Lương cơ bản không hợp lệ", http.StatusBadRequest)
			return
		}

		soNgayCong, err := strconv.Atoi(r.FormValue("so_ngay_cong"))
		if err != nil || soNgayCong < 0 {
			http.Error(w, "Số ngày công không hợp lệ", http.StatusBadRequest)
			return
		}

		phuCap, err := strconv.ParseFloat(r.FormValue("phu_cap"), 64)
		if err != nil {
			phuCap = 0
		}

		khauTru, err := strconv.ParseFloat(r.FormValue("khau_tru"), 64)
		if err != nil {
			khauTru = 0
		}

		ghiChu := r.FormValue("ghi_chu")

		// Kiểm tra nhân viên có tồn tại không
		var tonTaiNhanVien int

		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM nhanvien
			WHERE ma_nv = ?
		`, maNV).Scan(&tonTaiNhanVien)

		if err != nil {
			http.Error(w, "Lỗi kiểm tra nhân viên: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if tonTaiNhanVien == 0 {
			http.Error(w, "Nhân viên không tồn tại", http.StatusBadRequest)
			return
		}

		// Kiểm tra trùng bảng lương
		var daTonTai int

		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM bangluong
			WHERE ma_nv = ?
			  AND thang = ?
			  AND nam = ?
		`, maNV, thang, nam).Scan(&daTonTai)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if daTonTai > 0 {
			http.Error(
				w,
				"Bảng lương của nhân viên này trong tháng đã tồn tại",
				http.StatusConflict,
			)
			return
		}

		// =========================================================
		// TÍNH SỐ NGÀY CÔNG CHUẨN TRONG THÁNG
		// Chỉ tính thứ 2 -> thứ 6
		// =========================================================

		ngayDauThang := time.Date(
			nam,
			time.Month(thang),
			1,
			0, 0, 0, 0,
			time.Local,
		)

		ngayCuoiThang := ngayDauThang.AddDate(0, 1, -1)

		soNgayCongChuan := 0

		for ngay := ngayDauThang; !ngay.After(ngayCuoiThang); ngay = ngay.AddDate(0, 0, 1) {
			thu := ngay.Weekday()

			if thu == time.Saturday || thu == time.Sunday {
				continue
			}

			soNgayCongChuan++
		}

		if soNgayCongChuan == 0 {
			http.Error(w, "Không xác định được số ngày công chuẩn", http.StatusBadRequest)
			return
		}

		// Không cho số ngày công vượt quá số ngày công chuẩn
		if soNgayCong > soNgayCongChuan {
			soNgayCong = soNgayCongChuan
		}

		// =========================================================
		// TÍNH TỔNG LƯƠNG
		// =========================================================

		luongTheoNgay := luongCoBan / float64(soNgayCongChuan)

		tongLuong := luongTheoNgay*float64(soNgayCong) +
			phuCap -
			khauTru

		// =========================================================
		// LƯU BẢNG LƯƠNG
		// =========================================================

		_, err = db.Exec(`
			INSERT INTO bangluong
			(
				ma_nv,
				thang,
				nam,
				luong_co_ban,
				so_ngay_cong,
				phu_cap,
				khau_tru,
				tong_luong,
				ghi_chu
			)
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

			// ==============================
			// CHUYỂN DỮ LIỆU
			// ==============================

			thangValue, err := strconv.Atoi(thang)
			if err != nil || thangValue < 1 || thangValue > 12 {
				http.Error(w, "Tháng không hợp lệ", http.StatusBadRequest)
				return
			}

			namValue, err := strconv.Atoi(nam)
			if err != nil || namValue < 2000 {
				http.Error(w, "Năm không hợp lệ", http.StatusBadRequest)
				return
			}

			soNgayCongValue, err := strconv.Atoi(soNgayCong)
			if err != nil || soNgayCongValue < 0 {
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

			// ==============================
			// LẤY LƯƠNG CƠ BẢN
			// ==============================

			var luongCoBan float64

			err = db.QueryRow(`
				SELECT luong_co_ban
				FROM hopdong
				WHERE ma_nv = ?
				  AND nhom_hop_dong = 'lao_dong'
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

			// ==============================
			// TÍNH SỐ NGÀY CÔNG CHUẨN
			// THỨ 2 -> THỨ 6
			// ==============================

			ngayDauThang := time.Date(
				namValue,
				time.Month(thangValue),
				1,
				0, 0, 0, 0,
				time.Local,
			)

			ngayCuoiThang := ngayDauThang.AddDate(0, 1, -1)

			soNgayCongChuan := 0

			for ngay := ngayDauThang; !ngay.After(ngayCuoiThang); ngay = ngay.AddDate(0, 0, 1) {

				thu := ngay.Weekday()

				if thu == time.Saturday || thu == time.Sunday {
					continue
				}

				soNgayCongChuan++
			}

			if soNgayCongChuan == 0 {
				http.Error(
					w,
					"Không xác định được số ngày công chuẩn",
					http.StatusBadRequest,
				)
				return
			}

			// Không cho số ngày công vượt quá ngày công chuẩn
			if soNgayCongValue > soNgayCongChuan {
				soNgayCongValue = soNgayCongChuan
			}

			// ==============================
			// TÍNH LẠI TỔNG LƯƠNG
			// ==============================

			luongTheoNgay := luongCoBan / float64(soNgayCongChuan)

			tongLuong := luongTheoNgay*float64(soNgayCongValue) +
				phuCapValue -
				khauTruValue

			// ==============================
			// CẬP NHẬT BẢNG LƯƠNG
			// ==============================

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
			return
		}

		// ==============================
		// GET: HIỂN THỊ FORM SỬA
		// ==============================

		var bl models.BangLuong

		err = db.QueryRow(`
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
			http.Error(w, "Không tìm thấy bảng lương", http.StatusNotFound)
			return
		}

		tmpl, err := template.ParseFiles("templates/bangluong/sua.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, bl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
