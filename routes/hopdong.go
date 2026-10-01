package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"qlnv/models"
)

// Danh sách hợp đồng
func DanhSachHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, ma_hd, ma_nv, loai_hop_dong,
			       ngay_bat_dau, ngay_ket_thuc,
			       luong_co_ban, ghi_chu
			FROM hopdong
			ORDER BY id DESC
		`)

		if err != nil {
			http.Error(w, "Lỗi truy vấn database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var danhSach []models.HopDong

		for rows.Next() {
			var hd models.HopDong

			err := rows.Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.LoaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&hd.LuongCoBan,
				&hd.GhiChu,
			)

			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(), http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, hd)
		}

		tmpl := template.Must(template.ParseFiles(
			"templates/hopdong/list.html",
		))

		tmpl.Execute(w, danhSach)
	}
}

// Thêm hợp đồng
func ThemHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			tmpl := template.Must(template.ParseFiles(
				"templates/hopdong/them.html",
			))

			tmpl.Execute(w, nil)
			return
		}

		if r.Method == "POST" {
			maHD := r.FormValue("ma_hd")
			maNV := r.FormValue("ma_nv")
			loaiHopDong := r.FormValue("loai_hop_dong")
			ngayBatDau := r.FormValue("ngay_bat_dau")
			ngayKetThuc := r.FormValue("ngay_ket_thuc")
			luongCoBan := r.FormValue("luong_co_ban")
			ghiChu := r.FormValue("ghi_chu")

			// Kiểm tra mã nhân viên có tồn tại không
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
					template.ParseFiles("templates/hopdong/them.html"),
				)

				data := struct {
					Error string
				}{
					Error: "Mã nhân viên " + maNV + " không tồn tại!",
				}

				tmpl.Execute(w, data)
				return
			}

			// Nếu mã nhân viên tồn tại thì mới thêm hop dong
			_, err = db.Exec(`
					INSERT INTO hopdong
					(ma_hd, ma_nv, loai_hop_dong, ngay_bat_dau,
					ngay_ket_thuc, luong_co_ban, ghi_chu)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`,
				maHD,
				maNV,
				loaiHopDong,
				ngayBatDau,
				ngayKetThuc,
				luongCoBan,
				ghiChu,
			)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/hopdong", http.StatusSeeOther)
		}

		// 	maHD := r.FormValue("ma_hd")
		// 	maNV := r.FormValue("ma_nv")
		// 	loaiHopDong := r.FormValue("loai_hop_dong")
		// 	ngayBatDau := r.FormValue("ngay_bat_dau")
		// 	ngayKetThuc := r.FormValue("ngay_ket_thuc")
		// 	luongCoBan := r.FormValue("luong_co_ban")
		// 	ghiChu := r.FormValue("ghi_chu")

		// 	_, err := db.Exec(`
		// 		INSERT INTO hopdong
		// 		(ma_hd, ma_nv, loai_hop_dong, ngay_bat_dau,
		// 		 ngay_ket_thuc, luong_co_ban, ghi_chu)
		// 		VALUES (?, ?, ?, ?, ?, ?, ?)
		// 	`,
		// 		maHD,
		// 		maNV,
		// 		loaiHopDong,
		// 		ngayBatDau,
		// 		ngayKetThuc,
		// 		luongCoBan,
		// 		ghiChu,
		// 	)

		// 	if err != nil {
		// 		http.Error(w, "Lỗi thêm hợp đồng: "+err.Error(), http.StatusInternalServerError)
		// 		return
		// 	}

		// 	http.Redirect(w, r, "/hopdong", http.StatusSeeOther)
	}
}

// Sửa hợp đồng
func SuaHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.URL.Query().Get("id")

		if r.Method == http.MethodGet {

			var hd models.HopDong

			err := db.QueryRow(`
				SELECT id, ma_hd, ma_nv, loai_hop_dong,
				       ngay_bat_dau, ngay_ket_thuc,
				       luong_co_ban, ghi_chu
				FROM hopdong
				WHERE id = ?
			`, id).Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.LoaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&hd.LuongCoBan,
				&hd.GhiChu,
			)

			if err != nil {
				http.Error(w, "Không tìm thấy hợp đồng", http.StatusNotFound)
				return
			}

			tmpl := template.Must(template.ParseFiles(
				"templates/hopdong/sua.html",
			))

			tmpl.Execute(w, hd)
			return
		}

		maHD := r.FormValue("ma_hd")
		maNV := r.FormValue("ma_nv")
		loaiHopDong := r.FormValue("loai_hop_dong")
		ngayBatDau := r.FormValue("ngay_bat_dau")
		ngayKetThuc := r.FormValue("ngay_ket_thuc")
		luongCoBan := r.FormValue("luong_co_ban")
		ghiChu := r.FormValue("ghi_chu")

		_, err := db.Exec(`
			UPDATE hopdong
			SET ma_hd = ?,
			    ma_nv = ?,
			    loai_hop_dong = ?,
			    ngay_bat_dau = ?,
			    ngay_ket_thuc = ?,
			    luong_co_ban = ?,
			    ghi_chu = ?
			WHERE id = ?
		`,
			maHD,
			maNV,
			loaiHopDong,
			ngayBatDau,
			ngayKetThuc,
			luongCoBan,
			ghiChu,
			id,
		)

		if err != nil {
			http.Error(w, "Lỗi sửa hợp đồng: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/hopdong", http.StatusSeeOther)
	}
}

// Xóa hợp đồng
func XoaHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.URL.Query().Get("id")

		idHD, err := strconv.Atoi(id)

		if err != nil {
			http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			"DELETE FROM hopdong WHERE id = ?",
			idHD,
		)

		if err != nil {
			http.Error(w, "Lỗi xóa hợp đồng: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/hopdong", http.StatusSeeOther)
	}
}
