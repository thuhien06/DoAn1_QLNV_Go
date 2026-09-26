package routes

import (
	"database/sql"
	"html/template"
	"net/http"

	"qlnv/models"
)

// Hiển thị danh sách nhân viên
func DanhSachNhanVien(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Truy vấn danh sách nhân viên từ database
		rows, err := db.Query(`
			SELECT id, ma_nv, ho_ten, ngay_sinh, gioi_tinh,
			       so_dien_thoai, email, dia_chi, ngay_vao_lam,
			       phong_ban, chuc_vu
			FROM nhanvien
		`)

		// Kiểm tra lỗi truy vấn
		if err != nil {
			http.Error(w, "Lỗi truy vấn database", http.StatusInternalServerError)
			return
		}

		// Đóng kết quả truy vấn sau khi sử dụng
		defer rows.Close()

		// Tạo danh sách để lưu thông tin nhân viên
		var danhSach []models.NhanVien

		// Duyệt qua từng nhân viên
		for rows.Next() {
			var nv models.NhanVien

			// Đọc dữ liệu từ database vào biến nhân viên
			err := rows.Scan(
				&nv.ID,
				&nv.MaNV,
				&nv.HoTen,
				&nv.NgaySinh,
				&nv.GioiTinh,
				&nv.SoDienThoai,
				&nv.Email,
				&nv.DiaChi,
				&nv.NgayVaoLam,
				&nv.PhongBan,
				&nv.ChucVu,
			)

			// Kiểm tra lỗi khi đọc dữ liệu
			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu", http.StatusInternalServerError)
				return
			}

			// Thêm nhân viên vào danh sách
			danhSach = append(danhSach, nv)
		}

		// Đọc giao diện danh sách nhân viên
		tmpl := template.Must(template.ParseFiles(
			"templates/nhanvien/list.html",
		))

		// Hiển thị danh sách nhân viên lên giao diện
		tmpl.Execute(w, danhSach)
	}
}
