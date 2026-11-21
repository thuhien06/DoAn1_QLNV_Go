package routes

import (
	"database/sql"
	"html/template"
	"net/http"

	"qlnv/models"
)

func HopDongSapHetHan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
                COALESCE(trang_thai, '')
            FROM hopdong
            WHERE ngay_ket_thuc IS NOT NULL
              AND ngay_bat_dau <= CURDATE()
              AND ngay_ket_thuc >= CURDATE()
              AND ngay_ket_thuc <= DATE_ADD(CURDATE(), INTERVAL 30 DAY)
            ORDER BY ngay_ket_thuc ASC
        `

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Lỗi truy vấn hợp đồng", http.StatusInternalServerError)
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
				&hd.NhomHopDong,
				&hd.LoaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&hd.LuongCoBan,
				&hd.GhiChu,
				&hd.TrangThai,
			)

			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu hợp đồng", http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, hd)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc dữ liệu", http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(
			template.ParseFiles("templates/hopdong/sap_het_han.html"),
		)

		data := struct {
			DanhSach []models.HopDong
		}{
			DanhSach: danhSach,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị danh sách hợp đồng", http.StatusInternalServerError)
			return
		}
	}
}
