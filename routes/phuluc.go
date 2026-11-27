package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"qlnv/models"
)

func DanhSachPhuLuc(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		maPhuLuc := strings.TrimSpace(r.URL.Query().Get("ma_phu_luc"))
		maHD := strings.TrimSpace(r.URL.Query().Get("ma_hd"))
		maNV := strings.TrimSpace(r.URL.Query().Get("ma_nv"))
		nhomHopDong := strings.TrimSpace(r.URL.Query().Get("nhom_hop_dong"))
		noiDung := strings.TrimSpace(r.URL.Query().Get("noi_dung_dieu_chinh"))

		// Điều kiện tìm kiếm
		where := ` WHERE 1=1 `
		var args []interface{}

		if maPhuLuc != "" {
			where += ` AND pl.ma_phu_luc LIKE ? `
			args = append(args, "%"+maPhuLuc+"%")
		}

		if maHD != "" {
			where += ` AND hd.ma_hd LIKE ? `
			args = append(args, "%"+maHD+"%")
		}

		if maNV != "" {
			where += ` AND hd.ma_nv LIKE ? `
			args = append(args, "%"+maNV+"%")
		}

		if nhomHopDong != "" {
			where += ` AND pl.nhom_hop_dong = ? `
			args = append(args, nhomHopDong)
		}

		if noiDung != "" {
			where += ` AND pl.noi_dung_dieu_chinh LIKE ? `
			args = append(args, "%"+noiDung+"%")
		}

		// Đếm số phụ lục sau khi lọc
		var tongSo int

		countQuery := `
			SELECT COUNT(*)
			FROM phuluchopdong pl
			JOIN hopdong hd ON pl.id_hop_dong = hd.id
		` + where

		err := db.QueryRow(countQuery, args...).Scan(&tongSo)
		if err != nil {
			http.Error(w, "Lỗi đếm số phụ lục: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		// Lấy danh sách phụ lục
		query := `
			SELECT
				pl.id,
				COALESCE(pl.ma_phu_luc, ''),
				pl.id_hop_dong,
				COALESCE(hd.ma_hd, ''),
				COALESCE(hd.ma_nv, ''),
				COALESCE(pl.nhom_hop_dong, ''),
				COALESCE(pl.noi_dung_dieu_chinh, ''),
				COALESCE(DATE_FORMAT(pl.ngay_tao, '%d/%m/%Y'), '')
			FROM phuluchopdong pl
			JOIN hopdong hd ON pl.id_hop_dong = hd.id
		` + where + `
			ORDER BY pl.id DESC
		`

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Lỗi truy vấn danh sách phụ lục: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.PhuLucHopDong

		for rows.Next() {
			var pl models.PhuLucHopDong

			err := rows.Scan(
				&pl.ID,
				&pl.MaPhuLuc,
				&pl.IDHopDong,
				&pl.MaHD,
				&pl.MaNV,
				&pl.NhomHopDong,
				&pl.NoiDungDieuChinh,
				&pl.NgayTao,
			)

			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu phụ lục: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, pl)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc danh sách phụ lục: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		data := struct {
			TongSo      int
			MaPhuLuc    string
			MaHD        string
			MaNV        string
			NhomHopDong string
			NoiDung     string
			DanhSach    []models.PhuLucHopDong
		}{
			TongSo:      tongSo,
			MaPhuLuc:    maPhuLuc,
			MaHD:        maHD,
			MaNV:        maNV,
			NhomHopDong: nhomHopDong,
			NoiDung:     noiDung,
			DanhSach:    danhSach,
		}

		tmpl, err := template.ParseFiles(
			"templates/phuluc/list.html",
		)
		if err != nil {
			http.Error(w, "Lỗi tải giao diện phụ lục: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị danh sách phụ lục: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}
}

func ThemPhuLuc(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type HopDongGoc struct {
			ID          int
			MaHD        string
			MaNV        string
			NhomHopDong string
			LoaiHopDong string
			NgayBatDau  string
			NgayKetThuc string
		}

		if r.Method == http.MethodPost {
			nhomHopDong := strings.TrimSpace(r.FormValue("nhom_hop_dong"))
			noiDung := strings.TrimSpace(r.FormValue("noi_dung_dieu_chinh"))
			idHopDong := strings.TrimSpace(r.FormValue("id_hop_dong"))

			if nhomHopDong == "" || noiDung == "" || idHopDong == "" {
				http.Error(w, "Vui lòng nhập đầy đủ thông tin.", http.StatusBadRequest)
				return
			}

			// Kiểm tra hợp đồng gốc có tồn tại và đúng nhóm không
			var id int

			err := db.QueryRow(`
				SELECT id
				FROM hopdong
				WHERE id = ? AND nhom_hop_dong = ?
			`, idHopDong, nhomHopDong).Scan(&id)

			if err != nil {
				http.Error(w, "Hợp đồng gốc không tồn tại hoặc không đúng nhóm hợp đồng.",
					http.StatusBadRequest)
				return
			}

			http.Redirect(
				w,
				r,
				"/phuluc/tao?id_hop_dong="+url.QueryEscape(idHopDong)+
					"&noi_dung_dieu_chinh="+url.QueryEscape(noiDung),
				http.StatusSeeOther,
			)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Phương thức không được hỗ trợ.", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
			SELECT
				id,
				COALESCE(ma_hd, ''),
				COALESCE(ma_nv, ''),
				COALESCE(nhom_hop_dong, ''),
				COALESCE(loai_hop_dong, ''),
				COALESCE(DATE_FORMAT(ngay_bat_dau, '%d/%m/%Y'), ''),
				COALESCE(DATE_FORMAT(ngay_ket_thuc, '%d/%m/%Y'), '')
			FROM hopdong
			ORDER BY id DESC
		`)

		if err != nil {
			http.Error(w, "Lỗi lấy danh sách hợp đồng: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []HopDongGoc

		for rows.Next() {
			var hd HopDongGoc

			if err := rows.Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.NhomHopDong,
				&hd.LoaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
			); err != nil {
				http.Error(w, "Lỗi đọc hợp đồng: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, hd)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc danh sách hợp đồng: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		data := struct {
			DanhSach []HopDongGoc
		}{
			DanhSach: danhSach,
		}

		tmpl, err := template.ParseFiles("templates/phuluc/them.html")
		if err != nil {
			http.Error(w, "Lỗi tải giao diện: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị giao diện: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}
}
