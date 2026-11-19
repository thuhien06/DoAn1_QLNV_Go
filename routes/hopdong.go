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

func TrangChuHopDong(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Cấu trúc lưu số liệu thống kê hợp đồng
		type ThongKeHopDong struct {
			TongHopDong int
			DangHieuLuc int
			SapHetHan   int
			DaHetHan    int
		}

		var tk ThongKeHopDong

		// Lấy số liệu thống kê trực tiếp từ database
		err := db.QueryRow(`
            SELECT
                COUNT(*),

                COALESCE(SUM(
                    CASE
                        WHEN ngay_bat_dau <= CURDATE()
                        AND ngay_ket_thuc >= CURDATE()
                        THEN 1 ELSE 0
                    END
                ), 0),

                COALESCE(SUM(
                    CASE
                        WHEN ngay_bat_dau <= CURDATE()
                        AND ngay_ket_thuc BETWEEN CURDATE()
                        AND DATE_ADD(CURDATE(), INTERVAL 30 DAY)
                        THEN 1 ELSE 0
                    END
                ), 0),

                COALESCE(SUM(
                    CASE
                        WHEN ngay_ket_thuc < CURDATE()
                        THEN 1 ELSE 0
                    END
                ), 0)

            FROM hopdong
        `).Scan(
			&tk.TongHopDong,
			&tk.DangHieuLuc,
			&tk.SapHetHan,
			&tk.DaHetHan,
		)

		if err != nil {
			http.Error(w, "Lỗi thống kê hợp đồng: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(
			template.ParseFiles("templates/hopdong/index.html"),
		)

		if err := tmpl.Execute(w, tk); err != nil {
			http.Error(w, "Lỗi hiển thị trang quản lý hợp đồng",
				http.StatusInternalServerError)
		}
	}
}

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

			// Các trường có thể nhận NULL từ database
			var luongCoBan sql.NullFloat64
			var loaiHopDong sql.NullString
			var ghiChu sql.NullString

			err := rows.Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.NhomHopDong,
				&loaiHopDong,
				&hd.NgayBatDau,
				&hd.NgayKetThuc,
				&luongCoBan,
				&ghiChu,
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

			// Xử lý loại hợp đồng
			if loaiHopDong.Valid {
				hd.LoaiHopDong = loaiHopDong.String
			} else {
				hd.LoaiHopDong = ""
			}

			// Xử lý lương cơ bản
			if luongCoBan.Valid {
				hd.LuongCoBan = luongCoBan.Float64
			} else {
				hd.LuongCoBan = 0
			}

			// Xử lý ghi chú
			if ghiChu.Valid {
				hd.GhiChu = ghiChu.String
			} else {
				hd.GhiChu = ""
			}

			// Xác định trạng thái hợp đồng
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
			http.Error(w, "Lỗi hiển thị danh sách: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func taoMaHD(tx *sql.Tx) (string, error) {
	nam := time.Now().Year()
	tienTo := fmt.Sprintf("HD%d", nam)

	var soLuong int
	err := tx.QueryRow(`
        SELECT COUNT(*) FROM hopdong
        WHERE ma_hd LIKE ?
    `, tienTo+"%").Scan(&soLuong)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", tienTo, soLuong+1), nil
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

		// Hàm hiển thị lỗi trên form
		hienLoi := func(message string) {
			tmpl.Execute(w, struct {
				Error string
			}{Error: message})
		}

		// Lấy thông tin chung
		nhom := r.FormValue("nhom_hop_dong")
		ngayKy := r.FormValue("ngay_ky")
		ghiChu := r.FormValue("ghi_chu")

		if ngayKy == "" {
			hienLoi("Vui lòng nhập ngày ký!")
			return
		}

		if nhom != "lao_dong" &&
			nhom != "kinh_doanh" &&
			nhom != "van_hanh" {
			hienLoi("Nhóm hợp đồng không hợp lệ!")
			return
		}

		// Chuyển chuỗi rỗng thành NULL
		nullable := func(s string) interface{} {
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
			return s
		}

		// Các biến dùng chung cho hợp đồng
		var maNVDB interface{}
		var loaiDB, ngayBatDauDB, ngayKetThucDB interface{}
		var luongDB, phuCapDB interface{}
		var doiTacDB, nguoiDaiDienDB, noiDungDB interface{}
		var giaTriDB, dichVuDB, chiPhiDB interface{}

		ngayKyDB := ngayKy

		// Bắt đầu transaction
		tx, err := db.Begin()
		if err != nil {
			hienLoi("Không thể bắt đầu giao dịch: " + err.Error())
			return
		}
		defer tx.Rollback()
		// Sinh mã hợp đồng tự động
		maHD, err := taoMaHD(tx)
		if err != nil {
			hienLoi("Không thể tạo mã hợp đồng: " + err.Error())
			return
		}

		switch nhom {

		// =================================
		// HỢP ĐỒNG LAO ĐỘNG
		// =================================
		case "lao_dong":

			hoTen := strings.TrimSpace(r.FormValue("ho_ten"))

			if hoTen == "" {
				hienLoi("Vui lòng nhập họ tên người lao động!")
				return
			}

			// Thông tin cá nhân
			ngaySinh := nullable(r.FormValue("ngay_sinh"))
			gioiTinh := nullable(r.FormValue("gioi_tinh"))
			soDienThoai := nullable(r.FormValue("so_dien_thoai"))
			email := nullable(r.FormValue("email"))

			quocTich := nullable(r.FormValue("quoc_tich"))
			noiSinh := nullable(r.FormValue("noi_sinh"))
			cccd := nullable(r.FormValue("cccd"))
			ngayCapCCCD := nullable(r.FormValue("ngay_cap_cccd"))
			noiCapCCCD := nullable(r.FormValue("noi_cap_cccd"))

			diaChiThuongTru := nullable(r.FormValue("dia_chi_thuong_tru"))
			diaChiTamTru := nullable(r.FormValue("dia_chi_tam_tru"))

			// Dùng địa chỉ thường trú làm địa chỉ chung
			diaChi := diaChiThuongTru
			if diaChi == nil {
				diaChi = diaChiTamTru
			}

			ngayBatDau := nullable(r.FormValue("ngay_bat_dau"))
			phongBan := nullable(r.FormValue("bo_phan"))
			chucVu := nullable(r.FormValue("chuc_danh"))

			// Tạo mã tạm để INSERT nhân viên
			maNVTam := fmt.Sprintf("T%019d", time.Now().UnixNano())

			result, err := tx.Exec(`
    INSERT INTO nhanvien (
        ma_nv, ho_ten, ngay_sinh, gioi_tinh,
        so_dien_thoai, email, dia_chi, ngay_vao_lam,
        phong_ban, chuc_vu, quoc_tich, noi_sinh,
        cccd, ngay_cap_cccd, noi_cap_cccd,
        dia_chi_thuong_tru, dia_chi_tam_tru
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
				maNVTam, hoTen, ngaySinh, gioiTinh,
				soDienThoai, email, diaChi, ngayBatDau,
				phongBan, chucVu, quocTich, noiSinh,
				cccd, ngayCapCCCD, noiCapCCCD,
				diaChiThuongTru, diaChiTamTru,
			)

			if err != nil {
				hienLoi("Lỗi tạo hồ sơ nhân viên: " + err.Error())
				return
			}

			// Lấy ID tự tăng vừa được tạo
			id, err := result.LastInsertId()
			if err != nil {
				hienLoi("Không lấy được ID nhân viên: " + err.Error())
				return
			}

			// Tạo mã nhân viên từ ID
			maNV := fmt.Sprintf("NV%04d", id)

			// Cập nhật mã nhân viên chính thức
			_, err = tx.Exec(
				"UPDATE nhanvien SET ma_nv = ? WHERE id = ?",
				maNV, id,
			)

			if err != nil {
				hienLoi("Lỗi cập nhật mã nhân viên: " + err.Error())
				return
			}

			maNVDB = maNV

			// Thông tin hợp đồng lao động
			loaiDB = nullable(r.FormValue("loai_hop_dong"))
			ngayBatDauDB = ngayBatDau
			ngayKetThucDB = nullable(r.FormValue("ngay_ket_thuc"))
			luongDB = nullable(r.FormValue("luong_co_ban"))
			phuCapDB = nullable(r.FormValue("phu_cap"))
			noiDungDB = nullable(r.FormValue("noi_dung"))

		// =================================
		// HỢP ĐỒNG KINH DOANH
		// =================================
		case "kinh_doanh":

			doiTacDB = nullable(r.FormValue("doi_tac"))
			nguoiDaiDienDB = nullable(r.FormValue("nguoi_dai_dien"))
			giaTriDB = nullable(r.FormValue("gia_tri"))
			noiDungDB = nullable(r.FormValue("noi_dung_kinh_doanh"))

		// =================================
		// HỢP ĐỒNG VẬN HÀNH
		// =================================
		case "van_hanh":

			doiTacDB = nullable(r.FormValue("doi_tac_van_hanh"))
			dichVuDB = nullable(r.FormValue("dich_vu_thiet_bi"))
			chiPhiDB = nullable(r.FormValue("chi_phi"))
			noiDungDB = nullable(r.FormValue("noi_dung_van_hanh"))
		}

		// Tạo hợp đồng
		_, err = tx.Exec(`
            INSERT INTO hopdong (
                ma_hd, ma_nv, nhom_hop_dong, loai_hop_dong,
                ngay_bat_dau, ngay_ket_thuc,
                luong_co_ban, phu_cap, ghi_chu,
                doi_tac, nguoi_dai_dien, noi_dung,
                gia_tri, ngay_ky, dich_vu_thiet_bi, chi_phi
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
			maHD, maNVDB, nhom, loaiDB,
			ngayBatDauDB, ngayKetThucDB,
			luongDB, phuCapDB, nullable(ghiChu),
			doiTacDB, nguoiDaiDienDB, noiDungDB,
			giaTriDB, ngayKyDB, dichVuDB, chiPhiDB,
		)

		if err != nil {
			hienLoi("Lỗi thêm hợp đồng: " + err.Error())
			return
		}

		// Xác nhận transaction
		if err = tx.Commit(); err != nil {
			hienLoi("Lỗi lưu dữ liệu: " + err.Error())
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
		session, err := Store.Get(r, "qlnv-session")
		if err != nil {
			http.Error(w, "Không thể đọc phiên đăng nhập", http.StatusInternalServerError)
			return
		}

		vaiTro, _ := session.Values["vai_tro"].(string)
		maNV, _ := session.Values["ma_nv"].(string)

		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Thiếu mã hợp đồng", http.StatusBadRequest)
			return
		}

		var query string
		var args []interface{}

		switch vaiTro {
		case "admin", "hr":
			query = `
				SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				       loai_hop_dong, ngay_bat_dau, ngay_ket_thuc,
				       luong_co_ban, phu_cap, ghi_chu,
				       doi_tac, nguoi_dai_dien, noi_dung,
				       gia_tri, ngay_ky, dich_vu_thiet_bi, chi_phi
				FROM hopdong
				WHERE id = ?
			`
			args = []interface{}{id}

		case "nhanvien":
			if maNV == "" {
				http.Error(w, "Không tìm thấy mã nhân viên", http.StatusForbidden)
				return
			}

			query = `
				SELECT id, ma_hd, ma_nv, nhom_hop_dong,
				       loai_hop_dong, ngay_bat_dau, ngay_ket_thuc,
				       luong_co_ban, phu_cap, ghi_chu,
				       doi_tac, nguoi_dai_dien, noi_dung,
				       gia_tri, ngay_ky, dich_vu_thiet_bi, chi_phi
				FROM hopdong
				WHERE id = ? AND ma_nv = ?
			`
			args = []interface{}{id, maNV}

		default:
			http.Error(w, "Bạn không có quyền xem hợp đồng", http.StatusForbidden)
			return
		}

		var hd models.HopDong

		var maNVDB, nhomHopDong, loaiHopDong sql.NullString
		var ghiChu, doiTac, nguoiDaiDien sql.NullString
		var noiDung, dichVu sql.NullString
		var ngayBatDau, ngayKetThuc, ngayKy sql.NullTime
		var luong, phuCap, giaTri, chiPhi sql.NullFloat64

		err = db.QueryRow(query, args...).Scan(
			&hd.ID,
			&hd.MaHD,
			&maNVDB,
			&nhomHopDong,
			&loaiHopDong,
			&ngayBatDau,
			&ngayKetThuc,
			&luong,
			&phuCap,
			&ghiChu,
			&doiTac,
			&nguoiDaiDien,
			&noiDung,
			&giaTri,
			&ngayKy,
			&dichVu,
			&chiPhi,
		)

		if err != nil {
			fmt.Println("Lỗi truy vấn chi tiết hợp đồng:", err)
			http.Error(w, "Không tìm thấy hợp đồng hoặc bạn không có quyền xem", http.StatusNotFound)
			return
		}

		// Chuyển dữ liệu từ các biến Null vào model
		if maNVDB.Valid {
			hd.MaNV = maNVDB.String
		}
		if nhomHopDong.Valid {
			hd.NhomHopDong = nhomHopDong.String
		}
		if loaiHopDong.Valid {
			hd.LoaiHopDong = loaiHopDong.String
		}
		if ghiChu.Valid {
			hd.GhiChu = ghiChu.String
		}
		if doiTac.Valid {
			hd.DoiTac = &doiTac.String
		}
		if nguoiDaiDien.Valid {
			hd.NguoiDaiDien = &nguoiDaiDien.String
		}
		if noiDung.Valid {
			hd.NoiDung = &noiDung.String
		}
		if dichVu.Valid {
			hd.DichVuThietBi = &dichVu.String
		}
		if ngayBatDau.Valid {
			hd.NgayBatDau = &ngayBatDau.Time
		}
		if ngayKetThuc.Valid {
			hd.NgayKetThuc = &ngayKetThuc.Time
		}
		if ngayKy.Valid {
			hd.NgayKy = &ngayKy.Time
		}
		if luong.Valid {
			hd.LuongCoBan = luong.Float64
		}
		if phuCap.Valid {
			hd.PhuCap = &phuCap.Float64
		}
		if giaTri.Valid {
			hd.GiaTri = &giaTri.Float64
		}
		if chiPhi.Valid {
			hd.ChiPhi = &chiPhi.Float64
		}

		// Dữ liệu truyền sang template
		data := struct {
			models.HopDong
			NhanVien models.NhanVien
		}{
			HopDong: hd,
		}

		// Chỉ lấy thông tin nhân viên đối với hợp đồng lao động
		if hd.NhomHopDong == "lao_dong" {

			var nv models.NhanVien

			var ngaySinh, ngayCap sql.NullTime

			var quocTich, noiSinh, cccd sql.NullString
			var noiCap, diaChiThuongTru, diaChiTamTru sql.NullString
			var phongBan, chucVu sql.NullString

			err := db.QueryRow(`
		SELECT ho_ten, ngay_sinh, quoc_tich, noi_sinh,
		       cccd, ngay_cap_cccd, noi_cap_cccd,
		       dia_chi_thuong_tru, dia_chi_tam_tru,
		       phong_ban, chuc_vu
		FROM nhanvien
		WHERE ma_nv = ?
	`, hd.MaNV).Scan(
				&nv.HoTen,
				&ngaySinh,
				&quocTich,
				&noiSinh,
				&cccd,
				&ngayCap,
				&noiCap,
				&diaChiThuongTru,
				&diaChiTamTru,
				&phongBan,
				&chucVu,
			)

			if err != nil {
				fmt.Println("Lỗi lấy thông tin nhân viên:", err)
				http.Error(w, "Không tìm thấy thông tin nhân viên", http.StatusInternalServerError)
				return
			}

			if ngaySinh.Valid {
				nv.NgaySinh = &ngaySinh.Time
			}
			if ngayCap.Valid {
				nv.NgayCapCCCD = &ngayCap.Time
			}
			if quocTich.Valid {
				nv.QuocTich = quocTich.String
			}
			if noiSinh.Valid {
				nv.NoiSinh = noiSinh.String
			}
			if cccd.Valid {
				nv.CCCD = cccd.String
			}
			if noiCap.Valid {
				nv.NoiCapCCCD = noiCap.String
			}
			if diaChiThuongTru.Valid {
				nv.DiaChiThuongTru = diaChiThuongTru.String
			}
			if diaChiTamTru.Valid {
				nv.DiaChiTamTru = diaChiTamTru.String
			}
			if phongBan.Valid {
				nv.PhongBan = phongBan.String
			}
			if chucVu.Valid {
				nv.ChucVu = chucVu.String
			}

			data.NhanVien = nv
		}

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
			http.Error(w, "Không thể tải giao diện hợp đồng", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
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
