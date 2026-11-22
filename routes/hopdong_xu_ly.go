package routes

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"qlnv/models"
)

func XuLyHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Lấy ID hợp đồng
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id <= 0 {
			http.Error(w, "ID hợp đồng không hợp lệ", http.StatusBadRequest)
			return
		}

		// Xử lý lựa chọn của HR
		if r.Method == http.MethodPost {
			quyetDinh := strings.TrimSpace(r.FormValue("quyet_dinh"))

			var trangThaiXuLy string

			switch quyetDinh {
			case "tai_ky":
				trangThaiXuLy = "Chờ tái ký"

			case "ket_thuc":
				trangThaiXuLy = "Đã xác nhận kết thúc"

			default:
				http.Error(w, "Lựa chọn không hợp lệ", http.StatusBadRequest)
				return
			}

			_, err := db.Exec(`
				UPDATE hopdong
				SET trang_thai_xu_ly = ?
				WHERE id = ?
			`, trangThaiXuLy, id)

			if err != nil {
				http.Error(w, "Lỗi cập nhật quyết định xử lý", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/hopdong/sap-het-han", http.StatusSeeOther)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Phương thức không hợp lệ", http.StatusMethodNotAllowed)
			return
		}

		// Lấy thông tin hợp đồng
		var hd models.HopDong
		var trangThaiXuLy sql.NullString

		query := `
			SELECT
				id,
				ma_hd,
				COALESCE(ma_nv, ''),
				COALESCE(nhom_hop_dong, ''),
				COALESCE(loai_hop_dong, ''),
				ngay_bat_dau,
				ngay_ket_thuc,
				COALESCE(luong_co_ban, 0),
				COALESCE(ghi_chu, ''),
				COALESCE(trang_thai_xu_ly, 'Chưa xử lý')
			FROM hopdong
			WHERE id = ?
		`

		err = db.QueryRow(query, id).Scan(
			&hd.ID,
			&hd.MaHD,
			&hd.MaNV,
			&hd.NhomHopDong,
			&hd.LoaiHopDong,
			&hd.NgayBatDau,
			&hd.NgayKetThuc,
			&hd.LuongCoBan,
			&hd.GhiChu,
			&trangThaiXuLy,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Không tìm thấy hợp đồng", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "Lỗi lấy thông tin hợp đồng", http.StatusInternalServerError)
			return
		}

		hd.TrangThai = "Chưa xử lý"
		if trangThaiXuLy.Valid {
			hd.TrangThai = trangThaiXuLy.String
		}

		data := struct {
			HopDong models.HopDong
		}{
			HopDong: hd,
		}

		tmpl, err := template.ParseFiles("templates/hopdong/xu_ly.html")
		if err != nil {
			http.Error(w, "Lỗi tải giao diện xử lý hợp đồng", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị trang xử lý", http.StatusInternalServerError)
			return
		}
	}
}
