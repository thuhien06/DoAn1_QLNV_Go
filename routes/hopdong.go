package routes

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"qlnv/models"
	"time"
)

// Danh sách hợp đồng
func DanhSachHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		maHD := strings.TrimSpace(r.URL.Query().Get("ma_hd"))
		maNV := strings.TrimSpace(r.URL.Query().Get("ma_nv"))
		nhomHopDong := strings.TrimSpace(r.URL.Query().Get("nhom_hop_dong"))

		query := `
			SELECT id, ma_hd, ma_nv, nhom_hop_dong, loai_hop_dong,
				ngay_bat_dau, ngay_ket_thuc, luong_co_ban, ghi_chu,
				doi_tac, nguoi_dai_dien, noi_dung, gia_tri, ngay_ky,
				dich_vu_thiet_bi, chi_phi
			FROM hopdong
			WHERE ma_hd LIKE ?
			AND ma_nv LIKE ?
			AND (? = '' OR nhom_hop_dong = ?)
			ORDER BY id DESC
		`

		rows, err := db.Query(
			query,
			"%"+maHD+"%",
			"%"+maNV+"%",
			nhomHopDong,
			nhomHopDong,
		)
		if err != nil {
			http.Error(w, "Lỗi lấy danh sách hợp đồng: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var danhSach []models.HopDong

		for rows.Next() {
			var hd models.HopDong
			var luongCoBan sql.NullFloat64

			var loaiHopDong sql.NullString
			err := rows.Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.NhomHopDong,
				&loaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&luongCoBan,
				&hd.GhiChu,
				&hd.DoiTac,
				&hd.NguoiDaiDien,
				&hd.NoiDung,
				&hd.GiaTri,
				&hd.NgayKy,
				&hd.DichVuThietBi,
				&hd.ChiPhi,
			)
			if err != nil {
				http.Error(w, "Lỗi đọc dữ liệu: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if loaiHopDong.Valid {
				hd.LoaiHopDong = loaiHopDong.String
			} else {
				hd.LoaiHopDong = ""
			}

			if luongCoBan.Valid {
				hd.LuongCoBan = luongCoBan.Float64
			} else {
				hd.LuongCoBan = 0
			}

			if hd.NgayKetThuc == nil {
				hd.TrangThai = "Không xác định"
			} else {
				homNay := time.Now()
				homNay = time.Date(
					homNay.Year(), homNay.Month(), homNay.Day(),
					0, 0, 0, 0, homNay.Location(),
				)

				ngayKT := *hd.NgayKetThuc
				ngayKT = time.Date(
					ngayKT.Year(), ngayKT.Month(), ngayKT.Day(),
					0, 0, 0, 0, homNay.Location(),
				)

				soNgay := int(ngayKT.Sub(homNay).Hours() / 24)

				switch {
				case soNgay < 0:
					hd.TrangThai = "Đã hết hạn"
				case soNgay <= 30:
					hd.TrangThai = "Sắp hết hạn"
				default:
					hd.TrangThai = "Còn hạn"
				}
			}

			danhSach = append(danhSach, hd)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc danh sách hợp đồng: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			TongSo      int
			DanhSach    []models.HopDong
			MaHD        string
			MaNV        string
			NhomHopDong string
		}{
			TongSo:      len(danhSach),
			DanhSach:    danhSach,
			MaHD:        maHD,
			MaNV:        maNV,
			NhomHopDong: nhomHopDong,
		}

		tmpl := template.Must(
			template.ParseFiles("templates/hopdong/list.html"),
		)

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Lỗi hiển thị danh sách", http.StatusInternalServerError)
		}
	}
}

func ThemHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/hopdong/them.html"))

		if r.Method == http.MethodGet {
			tmpl.Execute(w, nil)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không hợp lệ", http.StatusMethodNotAllowed)
			return
		}

		maHD := r.FormValue("ma_hd")
		maNV := r.FormValue("ma_nv")
		nhom := r.FormValue("nhom_hop_dong")
		loai := r.FormValue("loai_hop_dong")
		ngayBatDau := r.FormValue("ngay_bat_dau")
		ngayKetThuc := r.FormValue("ngay_ket_thuc")
		ngayKy := r.FormValue("ngay_ky")
		luongCoBan := r.FormValue("luong_co_ban")
		ghiChu := r.FormValue("ghi_chu")

		doiTac := r.FormValue("doi_tac")
		nguoiDaiDien := r.FormValue("nguoi_dai_dien")
		noiDung := r.FormValue("noi_dung")
		giaTri := r.FormValue("gia_tri")
		dichVuThietBi := r.FormValue("dich_vu_thiet_bi")
		chiPhi := r.FormValue("chi_phi")

		if nhom != "lao_dong" &&
			nhom != "kinh_doanh" &&
			nhom != "van_hanh" {
			http.Error(w, "Nhóm hợp đồng không hợp lệ", http.StatusBadRequest)
			return
		}

		// Kiểm tra nhân viên phụ trách
		var count int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM nhanvien WHERE ma_nv = ?",
			maNV,
		).Scan(&count)

		if err != nil {
			http.Error(w, "Lỗi kiểm tra mã nhân viên: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		if count == 0 {
			tmpl.Execute(w, struct {
				Error string
			}{
				Error: "Mã nhân viên " + maNV + " không tồn tại!",
			})
			return
		}

		// Đổi trường rỗng thành NULL để tránh lỗi với DATE và DECIMAL
		nullable := func(s string) interface{} {
			if s == "" {
				return nil
			}
			return s
		}

		// Các trường dùng riêng cho từng nhóm
		var loaiDB, ngayBatDauDB, ngayKetThucDB interface{}
		var ngayKyDB, luongDB interface{}
		var doiTacDB, nguoiDaiDienDB, noiDungDB interface{}
		var giaTriDB, dichVuDB, chiPhiDB interface{}

		ngayBatDauDB = nullable(ngayBatDau)
		ngayKetThucDB = nullable(ngayKetThuc)
		ngayKyDB = nullable(ngayKy)
		loaiDB = nullable(loai)
		luongDB = nullable(luongCoBan)
		doiTacDB = nullable(doiTac)
		nguoiDaiDienDB = nullable(nguoiDaiDien)
		noiDungDB = nullable(noiDung)
		giaTriDB = nullable(giaTri)
		dichVuDB = nullable(dichVuThietBi)
		chiPhiDB = nullable(chiPhi)

		switch nhom {
		case "lao_dong":
			// Hợp đồng lao động không dùng thông tin đối tác.
			doiTacDB = nil
			nguoiDaiDienDB = nil
			noiDungDB = nil
			giaTriDB = nil
			ngayKyDB = nil
			dichVuDB = nil
			chiPhiDB = nil

		case "kinh_doanh":
			// Hợp đồng kinh doanh không dùng lương cơ bản.
			loaiDB = nil
			luongDB = nil
			dichVuDB = nil
			chiPhiDB = nil

		case "van_hanh":
			// Hợp đồng vận hành dùng chi phí, không dùng lương.
			loaiDB = nil
			luongDB = nil
			nguoiDaiDienDB = nil
			giaTriDB = nil
		}

		_, err = db.Exec(`
			INSERT INTO hopdong (
				ma_hd, ma_nv, nhom_hop_dong, loai_hop_dong,
				ngay_bat_dau, ngay_ket_thuc, luong_co_ban, ghi_chu,
				doi_tac, nguoi_dai_dien, noi_dung, gia_tri,
				ngay_ky, dich_vu_thiet_bi, chi_phi
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			maHD, maNV, nhom, loaiDB,
			ngayBatDauDB, ngayKetThucDB, luongDB, ghiChu,
			doiTacDB, nguoiDaiDienDB, noiDungDB, giaTriDB,
			ngayKyDB, dichVuDB, chiPhiDB,
		)

		if err != nil {
			http.Error(w, "Lỗi thêm hợp đồng: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/hopdong", http.StatusSeeOther)
	}
}

// Sửa hợp đồng
func SuaHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		if r.Method == http.MethodGet {
			var hd models.HopDong

			var nhomHopDong, loaiHopDong sql.NullString
			var luongCoBan sql.NullFloat64

			err := db.QueryRow(`
				SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				       loai_hop_dong, ngay_bat_dau, ngay_ket_thuc,
				       luong_co_ban, ghi_chu, doi_tac,
				       nguoi_dai_dien, noi_dung, gia_tri,
				       ngay_ky, dich_vu_thiet_bi, chi_phi
				FROM hopdong
				WHERE id = ?
			`, id).Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&nhomHopDong,
				&loaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&luongCoBan,
				&hd.GhiChu,
				&hd.DoiTac,
				&hd.NguoiDaiDien,
				&hd.NoiDung,
				&hd.GiaTri,
				&hd.NgayKy,
				&hd.DichVuThietBi,
				&hd.ChiPhi,
			)

			if err != nil {
				http.Error(w, "Không tìm thấy hợp đồng: "+err.Error(), http.StatusNotFound)
				return
			}

			hd.NhomHopDong = nhomHopDong.String
			hd.LoaiHopDong = loaiHopDong.String
			if luongCoBan.Valid {
				hd.LuongCoBan = luongCoBan.Float64
			}

			funcMap := template.FuncMap{
				"deref": func(v *float64) float64 {
					if v == nil {
						return 0
					}
					return *v
				},
			}

			tmpl := template.Must(
				template.New("sua.html").
					Funcs(funcMap).
					ParseFiles("templates/hopdong/sua.html"),
			)
			if err := tmpl.Execute(w, hd); err != nil {
				http.Error(w, "Lỗi hiển thị hợp đồng: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		maHD := r.FormValue("ma_hd")
		maNV := r.FormValue("ma_nv")
		nhom := r.FormValue("nhom_hop_dong")
		loai := r.FormValue("loai_hop_dong")
		ngayBatDau := r.FormValue("ngay_bat_dau")
		ngayKetThuc := r.FormValue("ngay_ket_thuc")
		luongCoBan := r.FormValue("luong_co_ban")
		ghiChu := r.FormValue("ghi_chu")

		doiTac := r.FormValue("doi_tac")
		nguoiDaiDien := r.FormValue("nguoi_dai_dien")
		noiDung := r.FormValue("noi_dung")
		giaTri := r.FormValue("gia_tri")
		ngayKy := r.FormValue("ngay_ky")
		dichVuThietBi := r.FormValue("dich_vu_thiet_bi")
		chiPhi := r.FormValue("chi_phi")

		if nhom != "lao_dong" &&
			nhom != "kinh_doanh" &&
			nhom != "van_hanh" {
			http.Error(w, "Nhóm hợp đồng không hợp lệ", http.StatusBadRequest)
			return
		}

		// Chỉ giữ các trường phù hợp với nhóm hợp đồng.
		if nhom == "lao_dong" {
			doiTac, nguoiDaiDien, noiDung = "", "", ""
			giaTri, ngayKy, dichVuThietBi, chiPhi = "", "", "", ""
		} else {
			loai, luongCoBan = "", ""
			if nhom == "kinh_doanh" {
				dichVuThietBi, chiPhi = "", ""
			} else {
				nguoiDaiDien, giaTri = "", ""
			}
		}

		// Chuyển các trường tùy chọn rỗng thành NULL.
		nullable := func(s string) interface{} {
			if s == "" {
				return nil
			}
			return s
		}

		_, err := db.Exec(`
			UPDATE hopdong
			SET ma_hd = ?,
			    ma_nv = ?,
			    nhom_hop_dong = ?,
			    loai_hop_dong = ?,
			    ngay_bat_dau = ?,
			    ngay_ket_thuc = ?,
			    luong_co_ban = ?,
			    ghi_chu = ?,
			    doi_tac = ?,
			    nguoi_dai_dien = ?,
			    noi_dung = ?,
			    gia_tri = ?,
			    ngay_ky = ?,
			    dich_vu_thiet_bi = ?,
			    chi_phi = ?
			WHERE id = ?
		`,
			maHD,
			maNV,
			nhom,
			nullable(loai),
			nullable(ngayBatDau),
			nullable(ngayKetThuc),
			nullable(luongCoBan),
			ghiChu,
			nullable(doiTac),
			nullable(nguoiDaiDien),
			nullable(noiDung),
			nullable(giaTri),
			nullable(ngayKy),
			nullable(dichVuThietBi),
			nullable(chiPhi),
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

// Chi tiết hợp đồng
func ChiTietHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Lấy thông tin đăng nhập
		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Error(w, "Không thể đọc phiên đăng nhập",
				http.StatusInternalServerError)
			return
		}

		vaiTro, _ := session.Values["vai_tro"].(string)
		maNV, _ := session.Values["ma_nv"].(string)

		// Lấy ID hợp đồng
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Thiếu mã hợp đồng", http.StatusBadRequest)
			return
		}

		// Xây dựng truy vấn theo vai trò
		var query string
		var args []interface{}

		switch vaiTro {
		case "admin", "hr":
			query = `
				SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				       loai_hop_dong, ngay_bat_dau, ngay_ket_thuc,
				       luong_co_ban, ghi_chu, doi_tac, nguoi_dai_dien,
				       noi_dung, gia_tri, ngay_ky,
				       dich_vu_thiet_bi, chi_phi
				FROM hopdong
				WHERE id = ?
			`
			args = []interface{}{id}

		case "nhanvien":
			if maNV == "" {
				http.Error(w, "Không tìm thấy mã nhân viên",
					http.StatusForbidden)
				return
			}

			query = `
				SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				       loai_hop_dong, ngay_bat_dau, ngay_ket_thuc,
				       luong_co_ban, ghi_chu, doi_tac, nguoi_dai_dien,
				       noi_dung, gia_tri, ngay_ky,
				       dich_vu_thiet_bi, chi_phi
				FROM hopdong
				WHERE id = ? AND ma_nv = ?
			`
			args = []interface{}{id, maNV}

		default:
			http.Error(w, "Bạn không có quyền xem hợp đồng",
				http.StatusForbidden)
			return
		}

		// Đọc dữ liệu hợp đồng
		var hd models.HopDong
		var nhomHopDong, loaiHopDong sql.NullString
		var luongCoBan, giaTri, chiPhi sql.NullFloat64

		err = db.QueryRow(query, args...).Scan(
			&hd.ID,
			&hd.MaHD,
			&hd.MaNV,
			&nhomHopDong,
			&loaiHopDong,
			&hd.NgayBatDau,
			&hd.NgayKetThuc,
			&luongCoBan,
			&hd.GhiChu,
			&hd.DoiTac,
			&hd.NguoiDaiDien,
			&hd.NoiDung,
			&giaTri,
			&hd.NgayKy,
			&hd.DichVuThietBi,
			&chiPhi,
		)

		if err != nil {
			http.Error(w,
				"Không tìm thấy hợp đồng hoặc bạn không có quyền xem",
				http.StatusNotFound)
			return
		}

		// Chuyển các giá trị NULL về giá trị mặc định
		if nhomHopDong.Valid {
			hd.NhomHopDong = nhomHopDong.String
		}
		if loaiHopDong.Valid {
			hd.LoaiHopDong = loaiHopDong.String
		}
		if luongCoBan.Valid {
			hd.LuongCoBan = luongCoBan.Float64
		}
		if giaTri.Valid {
			hd.GiaTri = &giaTri.Float64
		}
		if chiPhi.Valid {
			hd.ChiPhi = &chiPhi.Float64
		}

		// Hàm hỗ trợ hiển thị các trường con trỏ float64
		funcMap := template.FuncMap{
			"deref": func(v *float64) float64 {
				if v == nil {
					return 0
				}
				return *v
			},
		}

		tmpl, err := template.New("chitiet.html").
			Funcs(funcMap).
			ParseFiles("templates/hopdong/chitiet.html")

		if err != nil {
			fmt.Println("Lỗi đọc template chi tiết hợp đồng:", err)
			http.Error(w, "Không thể tải giao diện hợp đồng",
				http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, hd); err != nil {
			fmt.Println("Lỗi hiển thị chi tiết hợp đồng:", err)
		}
	}
}

// HopDongCaNhan: nhân viên chỉ xem hợp đồng của mình
func HopDongCaNhan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Error(w, "Không thể lấy thông tin đăng nhập", http.StatusUnauthorized)
			return
		}

		maNV, ok := session.Values["ma_nv"].(string)
		if !ok || maNV == "" {
			http.Error(w, "Tài khoản chưa được liên kết với nhân viên", http.StatusForbidden)
			return
		}

		rows, err := db.Query(`
			SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				COALESCE(loai_hop_dong, ''),
				ngay_bat_dau, ngay_ket_thuc,
				COALESCE(luong_co_ban, 0),
				COALESCE(ghi_chu, '')
			FROM hopdong
			WHERE ma_nv = ?
			ORDER BY id DESC
		`, maNV)
		if err != nil {
			http.Error(w, "Không thể tải danh sách hợp đồng", http.StatusInternalServerError)
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
			)
			if err != nil {
				fmt.Println("Lỗi đọc dữ liệu hợp đồng:", err)
				http.Error(w, "Không thể đọc dữ liệu hợp đồng", http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, hd)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi khi đọc danh sách hợp đồng", http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(
			template.ParseFiles("templates/hopdong/ca-nhan.html"),
		)

		err = tmpl.Execute(w, danhSach)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

