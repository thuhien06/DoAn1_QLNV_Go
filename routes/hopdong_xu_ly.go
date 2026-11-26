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

		// Kiểm tra đăng nhập
		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Redirect(w, r, "/dangnhap", http.StatusSeeOther)
			return
		}

		vaiTro, _ := session.Values["vai_tro"].(string)

		if vaiTro != "admin" && vaiTro != "hr" {
			http.Error(w, "Bạn không có quyền xử lý hợp đồng",
				http.StatusForbidden)
			return
		}

		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id <= 0 {
			http.Error(w, "ID hợp đồng không hợp lệ",
				http.StatusBadRequest)
			return
		}

		// Xử lý quyết định
		if r.Method == http.MethodPost {

			if err := r.ParseForm(); err != nil {
				http.Error(w, "Dữ liệu gửi lên không hợp lệ",
					http.StatusBadRequest)
				return
			}

			trangNguon := r.FormValue("trang_nguon")
			quyetDinh := strings.TrimSpace(r.FormValue("quyet_dinh"))

			// Admin chỉ được thay đổi quyết định tại trang sắp hết hạn
			if vaiTro == "admin" && trangNguon != "sap_het_han" {
				http.Error(w, "Không có quyền chỉnh sửa tại trang này",
					http.StatusForbidden)
				return
			}

			var trangThaiXuLy string

			switch quyetDinh {
			case "tai_ky":
				trangThaiXuLy = "Chờ tái ký"

			case "ket_thuc":
				trangThaiXuLy = "Đã xác nhận kết thúc"

			default:
				http.Error(w, "Lựa chọn không hợp lệ",
					http.StatusBadRequest)
				return
			}

			var result sql.Result

			if vaiTro == "admin" {

				// Admin được xử lý mới hoặc thay đổi quyết định cũ.
				// Không được thay đổi hợp đồng đã tái ký.
				result, err = db.Exec(`
					UPDATE hopdong
					SET trang_thai_xu_ly = ?
					WHERE id = ?
					  AND ngay_ket_thuc IS NOT NULL
					  AND ngay_ket_thuc >= CURDATE()
					  AND ngay_ket_thuc <= DATE_ADD(CURDATE(), INTERVAL 30 DAY)
					  AND COALESCE(trang_thai_xu_ly, 'Chưa xử lý')
					      IN (
					          'Chưa xử lý',
					          'Chờ tái ký',
					          'Đã xác nhận kết thúc'
					      )
				`, trangThaiXuLy, id)

			} else {

				// HR chỉ được xử lý lần đầu.
				result, err = db.Exec(`
					UPDATE hopdong
					SET trang_thai_xu_ly = ?
					WHERE id = ?
					  AND COALESCE(trang_thai_xu_ly, 'Chưa xử lý')
					      = 'Chưa xử lý'
				`, trangThaiXuLy, id)
			}

			if err != nil {
				http.Error(w, "Lỗi cập nhật quyết định xử lý",
					http.StatusInternalServerError)
				return
			}

			affected, err := result.RowsAffected()
			if err != nil {
				http.Error(w, "Không thể kiểm tra kết quả cập nhật",
					http.StatusInternalServerError)
				return
			}

			// Kiểm tra trường hợp admin chọn lại đúng quyết định cũ.
			if affected == 0 && vaiTro == "admin" {
				var trangThaiHienTai string

				err = db.QueryRow(`
					SELECT COALESCE(trang_thai_xu_ly, 'Chưa xử lý')
					FROM hopdong
					WHERE id = ?
					  AND ngay_ket_thuc IS NOT NULL
					  AND ngay_ket_thuc >= CURDATE()
					  AND ngay_ket_thuc <= DATE_ADD(CURDATE(), INTERVAL 30 DAY)
				`, id).Scan(&trangThaiHienTai)

				if err == nil && trangThaiHienTai == trangThaiXuLy {
					affected = 1
				}
			}

			if affected != 1 {
				http.Error(w,
					"Hợp đồng không tồn tại, không đủ điều kiện xử lý hoặc đã tái ký",
					http.StatusConflict)
				return
			}

			if trangNguon == "lich_su" {
				http.Redirect(w, r, "/hopdong/lich-su",
					http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/hopdong/sap-het-han",
					http.StatusSeeOther)
			}
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Phương thức không hợp lệ",
				http.StatusMethodNotAllowed)
			return
		}

		trangNguon := r.URL.Query().Get("trang_nguon")

		// Admin chỉ được mở chức năng chỉnh sửa từ danh sách sắp hết hạn.
		if vaiTro == "admin" && trangNguon != "sap_het_han" {
			http.Error(w, "Không có quyền truy cập chức năng này",
				http.StatusForbidden)
			return
		}

		// Lấy thông tin hợp đồng
		var hd models.HopDong

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
			&hd.TrangThai,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Không tìm thấy hợp đồng",
				http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "Lỗi lấy thông tin hợp đồng",
				http.StatusInternalServerError)
			return
		}

		// Xác định quyền thao tác
		coTheXuLy := hd.TrangThai == "Chưa xử lý"

		coTheSuaQuyetDinh := vaiTro == "admin" &&
			trangNguon == "sap_het_han" &&
			(hd.TrangThai == "Chờ tái ký" ||
				hd.TrangThai == "Đã xác nhận kết thúc")

		coTheTaiKy := hd.TrangThai == "Chờ tái ký"

		data := struct {
			HopDong           models.HopDong
			TrangNguon        string
			VaiTro            string
			CoTheXuLy         bool
			CoTheSuaQuyetDinh bool
			CoTheTaiKy        bool
		}{
			HopDong:           hd,
			TrangNguon:        trangNguon,
			VaiTro:            vaiTro,
			CoTheXuLy:         coTheXuLy,
			CoTheSuaQuyetDinh: coTheSuaQuyetDinh,
			CoTheTaiKy:        coTheTaiKy,
		}

		tmpl, err := template.ParseFiles(
			"templates/hopdong/xu_ly.html",
		)
		if err != nil {
			http.Error(w, "Lỗi tải giao diện xử lý hợp đồng",
				http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị trang xử lý",
				http.StatusInternalServerError)
			return
		}
	}
}
