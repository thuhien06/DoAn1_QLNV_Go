package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qlnv/models"
)

func TaiKyHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id <= 0 {
			http.Error(w, "Mã hợp đồng không hợp lệ", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			var hd models.HopDong

			query := `
				SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				       loai_hop_dong, ngay_bat_dau, ngay_ket_thuc,
				       ngay_ky, luong_co_ban, phu_cap,
				       COALESCE(noi_dung, ''),
				       COALESCE(ghi_chu, ''),
				       COALESCE(trang_thai_xu_ly, 'Chưa xử lý')
				FROM hopdong
				WHERE id = ?
			`

			err := db.QueryRow(query, id).Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.NhomHopDong,
				&hd.LoaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&hd.NgayKy,
				&hd.LuongCoBan,
				&hd.PhuCap,
				&hd.NoiDung,
				&hd.GhiChu,
				&hd.TrangThaiXuLy,
			)

			if err != nil {
				http.Error(w, "Không tìm thấy hợp đồng", http.StatusNotFound)
				return
			}

			if hd.NhomHopDong != "lao_dong" {
				http.Error(w, "Hiện chỉ hỗ trợ tái ký hợp đồng lao động", http.StatusBadRequest)
				return
			}

			if hd.TrangThaiXuLy != "Chờ tái ký" {
				http.Error(w, "Hợp đồng chưa được xác nhận tái ký", http.StatusBadRequest)
				return
			}

			tmpl := template.Must(
				template.ParseFiles("templates/hopdong/tai_ky.html"),
			)

			trangNguon := r.URL.Query().Get("trang_nguon")

			data := struct {
				HopDong    models.HopDong
				NgayKy     string
				TrangNguon string
			}{
				HopDong:    hd,
				NgayKy:     time.Now().Format("2006-01-02"),
				TrangNguon: trangNguon,
			}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Lỗi hiển thị form tái ký", http.StatusInternalServerError)
			}
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Dữ liệu gửi lên không hợp lệ", http.StatusBadRequest)
			return
		}
		trangNguon := r.FormValue("trang_nguon")

		loaiHopDong := strings.TrimSpace(r.FormValue("loai_hop_dong"))
		ngayBatDau := strings.TrimSpace(r.FormValue("ngay_bat_dau"))
		ngayKetThuc := strings.TrimSpace(r.FormValue("ngay_ket_thuc"))
		ngayKy := strings.TrimSpace(r.FormValue("ngay_ky"))
		noiDung := strings.TrimSpace(r.FormValue("noi_dung"))
		ghiChu := strings.TrimSpace(r.FormValue("ghi_chu"))

		if loaiHopDong == "" || ngayBatDau == "" || ngayKy == "" {
			http.Error(w, "Vui lòng nhập đầy đủ thông tin bắt buộc", http.StatusBadRequest)
			return
		}

		ngayBatDauParsed, err := time.Parse("2006-01-02", ngayBatDau)
		if err != nil {
			http.Error(w, "Ngày bắt đầu không hợp lệ", http.StatusBadRequest)
			return
		}

		if _, err := time.Parse("2006-01-02", ngayKy); err != nil {
			http.Error(w, "Ngày ký không hợp lệ", http.StatusBadRequest)
			return
		}

		var ketThuc interface{}

		if loaiHopDong == "Không xác định thời hạn" {
			ketThuc = nil
		} else {
			if ngayKetThuc == "" {
				http.Error(w, "Vui lòng nhập ngày kết thúc", http.StatusBadRequest)
				return
			}

			ngayKetThucParsed, err := time.Parse("2006-01-02", ngayKetThuc)
			if err != nil {
				http.Error(w, "Ngày kết thúc không hợp lệ", http.StatusBadRequest)
				return
			}

			if ngayKetThucParsed.Before(ngayBatDauParsed) {
				http.Error(w, "Ngày kết thúc phải sau hoặc bằng ngày bắt đầu", http.StatusBadRequest)
				return
			}

			ketThuc = ngayKetThuc

			ketThuc = ngayKetThuc
		}

		luongCoBan, err := strconv.ParseFloat(r.FormValue("luong_co_ban"), 64)
		if err != nil || luongCoBan < 0 {
			http.Error(w, "Lương cơ bản không hợp lệ", http.StatusBadRequest)
			return
		}

		var phuCap interface{}
		if strings.TrimSpace(r.FormValue("phu_cap")) != "" {
			value, err := strconv.ParseFloat(r.FormValue("phu_cap"), 64)
			if err != nil || value < 0 {
				http.Error(w, "Phụ cấp không hợp lệ", http.StatusBadRequest)
				return
			}
			phuCap = value
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Không thể bắt đầu giao dịch", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var maNV, nhomHopDong, trangThaiXuLy string

		err = tx.QueryRow(`
			SELECT ma_nv, nhom_hop_dong, trang_thai_xu_ly
			FROM hopdong
			WHERE id = ?
			FOR UPDATE
		`, id).Scan(&maNV, &nhomHopDong, &trangThaiXuLy)

		if err != nil {
			http.Error(w, "Không tìm thấy hợp đồng cũ", http.StatusNotFound)
			return
		}

		if nhomHopDong != "lao_dong" || trangThaiXuLy != "Chờ tái ký" {
			http.Error(w, "Hợp đồng không ở trạng thái chờ tái ký", http.StatusBadRequest)
			return
		}

		maHD, err := taoMaHD(tx)
		if err != nil {
			http.Error(w, "Không thể tạo mã hợp đồng mới", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
			INSERT INTO hopdong (
				ma_hd, ma_nv, nhom_hop_dong, loai_hop_dong,
				ngay_bat_dau, ngay_ket_thuc, ngay_ky,
				luong_co_ban, phu_cap, noi_dung, ghi_chu,
				trang_thai_xu_ly
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'Chưa xử lý')
		`,
			maHD,
			maNV,
			nhomHopDong,
			loaiHopDong,
			ngayBatDau,
			ketThuc,
			ngayKy,
			luongCoBan,
			phuCap,
			noiDung,
			ghiChu,
		)

		if err != nil {
			http.Error(w, "Lỗi tạo hợp đồng mới: "+err.Error(), http.StatusInternalServerError)
			return
		}

		result, err := tx.Exec(`
			UPDATE hopdong
			SET trang_thai_xu_ly = 'Đã tái ký'
			WHERE id = ? AND trang_thai_xu_ly = 'Chờ tái ký'
		`, id)

		if err != nil {
			http.Error(w, "Lỗi cập nhật hợp đồng cũ", http.StatusInternalServerError)
			return
		}

		affected, err := result.RowsAffected()
		if err != nil || affected != 1 {
			http.Error(w, "Không thể xác nhận tái ký hợp đồng cũ", http.StatusConflict)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Không thể lưu hợp đồng tái ký", http.StatusInternalServerError)
			return
		}

		if trangNguon == "lich_su" {
			http.Redirect(w, r, "/hopdong/lich-su", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/hopdong/sap-het-han", http.StatusSeeOther)
		}
	}
}
