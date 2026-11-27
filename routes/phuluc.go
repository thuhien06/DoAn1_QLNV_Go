package routes

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

		if r.Method == http.MethodPost {
			nhomHopDong := strings.TrimSpace(r.FormValue("nhom_hop_dong"))
			noiDung := strings.TrimSpace(r.FormValue("noi_dung_dieu_chinh"))
			idHopDong := strings.TrimSpace(r.FormValue("id_hop_dong"))

			if nhomHopDong == "" || noiDung == "" || idHopDong == "" {
				http.Error(w, "Vui lòng nhập đầy đủ thông tin.", http.StatusBadRequest)
				return
			}

			// Kiểm tra hợp đồng gốc có tồn tại, đúng nhóm và đang hiệu lực
			var hopDongTonTai int

			err := db.QueryRow(`
				SELECT COUNT(*)
				FROM hopdong
				WHERE id = ?
				  AND nhom_hop_dong = ?
				  AND ngay_bat_dau <= CURDATE()
				  AND (ngay_ket_thuc IS NULL OR ngay_ket_thuc >= CURDATE())
			`, idHopDong, nhomHopDong).Scan(&hopDongTonTai)

			if err != nil {
				http.Error(w, "Lỗi kiểm tra hợp đồng gốc: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			if hopDongTonTai == 0 {
				http.Error(w, "Hợp đồng gốc không tồn tại, không đúng nhóm hoặc đã hết hiệu lực.",
					http.StatusBadRequest)
				return
			}

			// Chuyển sang bước nhập nội dung phụ lục
			params := url.Values{}
			params.Set("id_hop_dong", idHopDong)
			params.Set("noi_dung_dieu_chinh", noiDung)

			http.Redirect(
				w,
				r,
				"/phuluc/tao?"+params.Encode(),
				http.StatusSeeOther,
			)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Phương thức không được hỗ trợ.",
				http.StatusMethodNotAllowed)
			return
		}

		// Không cần tải toàn bộ hợp đồng ở đây.
		// Danh sách hợp đồng sẽ được tìm bằng AJAX qua TimHopDongGoc.
		tmpl, err := template.ParseFiles("templates/phuluc/them.html")
		if err != nil {
			http.Error(w, "Lỗi tải giao diện: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, "Lỗi hiển thị giao diện: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}
}

type HopDongGoiY struct {
	ID          int    `json:"id"`
	MaHD        string `json:"ma_hd"`
	MaNV        string `json:"ma_nv"`
	LoaiHopDong string `json:"loai_hop_dong"`
}

func TimHopDongGoc(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Phương thức không hợp lệ",
				http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		nhom := strings.TrimSpace(r.URL.Query().Get("nhom_hop_dong"))
		maHD := strings.TrimSpace(r.URL.Query().Get("ma_hd"))

		if nhom == "" {
			json.NewEncoder(w).Encode([]HopDongGoiY{})
			return
		}

		// Chỉ tìm hợp đồng đã bắt đầu và chưa hết hạn
		query := `
			SELECT
				id,
				COALESCE(ma_hd, ''),
				COALESCE(ma_nv, ''),
				COALESCE(loai_hop_dong, '')
			FROM hopdong
			WHERE nhom_hop_dong = ?
			  AND ngay_bat_dau <= CURDATE()
			  AND (ngay_ket_thuc IS NULL OR ngay_ket_thuc >= CURDATE())
			  AND ma_hd LIKE ?
			ORDER BY ma_hd
			LIMIT 20
		`

		rows, err := db.Query(query, nhom, "%"+maHD+"%")
		if err != nil {
			http.Error(w, "Không thể tìm hợp đồng",
				http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		danhSach := []HopDongGoiY{}

		for rows.Next() {
			var hd HopDongGoiY

			if err := rows.Scan(
				&hd.ID,
				&hd.MaHD,
				&hd.MaNV,
				&hd.LoaiHopDong,
			); err != nil {
				http.Error(w, "Lỗi đọc dữ liệu hợp đồng",
					http.StatusInternalServerError)
				return
			}

			danhSach = append(danhSach, hd)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Lỗi đọc danh sách hợp đồng",
				http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(danhSach)
	}
}

func TaoPhuLuc(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Các nội dung được phép điều chỉnh theo từng nhóm
		noiDungTheoNhom := map[string]map[string]string{
			"lao_dong": {
				"dieu_chinh_luong":     "Điều chỉnh mức lương",
				"dieu_chinh_phu_cap":   "Điều chỉnh phụ cấp",
				"dieu_chinh_thoi_han":  "Điều chỉnh thời hạn hợp đồng",
				"dieu_chinh_cong_viec": "Điều chỉnh công việc, chức danh",
				"dieu_chinh_dia_diem":  "Điều chỉnh địa điểm làm việc",
				"noi_dung_khac":        "Nội dung khác",
			},
			"kinh_doanh": {
				"dieu_chinh_gia_tri":    "Điều chỉnh giá trị hợp đồng",
				"dieu_chinh_thanh_toan": "Điều chỉnh phương thức thanh toán",
				"dieu_chinh_thoi_han":   "Điều chỉnh thời hạn hợp đồng",
				"dieu_chinh_noi_dung":   "Điều chỉnh nội dung công việc",
				"noi_dung_khac":         "Nội dung khác",
			},
			"van_hanh": {
				"dieu_chinh_dich_vu":     "Điều chỉnh dịch vụ, thiết bị",
				"dieu_chinh_chi_phi":     "Điều chỉnh chi phí vận hành",
				"dieu_chinh_thoi_han":    "Điều chỉnh thời hạn hợp đồng",
				"dieu_chinh_trach_nhiem": "Điều chỉnh trách nhiệm các bên",
				"noi_dung_khac":          "Nội dung khác",
			},
		}

		if r.Method == http.MethodPost {
			idHopDong := strings.TrimSpace(r.FormValue("id_hop_dong"))
			noiDung := strings.TrimSpace(r.FormValue("noi_dung_dieu_chinh"))
			noiDungCu := strings.TrimSpace(r.FormValue("noi_dung_cu"))
			noiDungMoi := strings.TrimSpace(r.FormValue("noi_dung_moi"))
			ngayApDung := strings.TrimSpace(r.FormValue("ngay_ap_dung"))
			lyDo := strings.TrimSpace(r.FormValue("ly_do"))

			if idHopDong == "" || noiDung == "" ||
				noiDungCu == "" || noiDungMoi == "" ||
				ngayApDung == "" || lyDo == "" {
				http.Error(w, "Vui lòng nhập đầy đủ thông tin phụ lục.",
					http.StatusBadRequest)
				return
			}

			// Kiểm tra ID hợp đồng
			id, err := strconv.Atoi(idHopDong)
			if err != nil || id <= 0 {
				http.Error(w, "Mã hợp đồng không hợp lệ.",
					http.StatusBadRequest)
				return
			}

			// Lấy nhóm hợp đồng và kiểm tra hiệu lực
			var nhomHopDong string

			err = db.QueryRow(`
				SELECT nhom_hop_dong
				FROM hopdong
				WHERE id = ?
				  AND ngay_bat_dau <= CURDATE()
				  AND (ngay_ket_thuc IS NULL OR ngay_ket_thuc >= CURDATE())
			`, id).Scan(&nhomHopDong)

			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Hợp đồng gốc không tồn tại hoặc đã hết hiệu lực.",
						http.StatusBadRequest)
					return
				}

				http.Error(w, "Lỗi kiểm tra hợp đồng: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			// Kiểm tra nội dung điều chỉnh có thuộc nhóm hợp đồng không
			tenNoiDung, hopLe := noiDungTheoNhom[nhomHopDong][noiDung]
			if !hopLe {
				http.Error(w, "Nội dung điều chỉnh không hợp lệ.",
					http.StatusBadRequest)
				return
			}

			// Sinh mã phụ lục
			maPhuLuc := "PL" + strconv.FormatInt(time.Now().UnixNano(), 10)

			// Lưu phụ lục
			_, err = db.Exec(`
				INSERT INTO phuluchopdong (
					ma_phu_luc,
					id_hop_dong,
					nhom_hop_dong,
					noi_dung_dieu_chinh,
					noi_dung_cu,
					noi_dung_moi,
					ngay_ap_dung,
					ly_do
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`,
				maPhuLuc,
				id,
				nhomHopDong,
				tenNoiDung,
				noiDungCu,
				noiDungMoi,
				ngayApDung,
				lyDo,
			)

			if err != nil {
				http.Error(w, "Lỗi lưu phụ lục: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/phuluc", http.StatusSeeOther)
			return
		}

		if r.Method == http.MethodPost {

			idHopDong := strings.TrimSpace(r.FormValue("id_hop_dong"))
			noiDung := strings.TrimSpace(r.FormValue("noi_dung_dieu_chinh"))
			noiDungCu := strings.TrimSpace(r.FormValue("noi_dung_cu"))
			noiDungMoi := strings.TrimSpace(r.FormValue("noi_dung_moi"))
			ngayApDung := strings.TrimSpace(r.FormValue("ngay_ap_dung"))
			lyDo := strings.TrimSpace(r.FormValue("ly_do"))

			if idHopDong == "" || noiDung == "" ||
				noiDungCu == "" || noiDungMoi == "" ||
				ngayApDung == "" || lyDo == "" {

				http.Error(w, "Vui lòng nhập đầy đủ thông tin phụ lục.",
					http.StatusBadRequest)
				return
			}

			// Kiểm tra ID hợp đồng
			id, err := strconv.Atoi(idHopDong)
			if err != nil || id <= 0 {
				http.Error(w, "Mã hợp đồng không hợp lệ.",
					http.StatusBadRequest)
				return
			}

			// Kiểm tra định dạng ngày
			if _, err := time.Parse("2006-01-02", ngayApDung); err != nil {
				http.Error(w, "Ngày áp dụng không hợp lệ.",
					http.StatusBadRequest)
				return
			}

			// Kiểm tra hợp đồng gốc còn hiệu lực
			var nhomHopDong string

			err = db.QueryRow(`
        SELECT nhom_hop_dong
        FROM hopdong
        WHERE id = ?
          AND ngay_bat_dau <= CURDATE()
          AND (ngay_ket_thuc IS NULL OR ngay_ket_thuc >= CURDATE())
    `, id).Scan(&nhomHopDong)

			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Hợp đồng gốc không tồn tại hoặc đã hết hiệu lực.",
						http.StatusBadRequest)
					return
				}

				http.Error(w, "Lỗi kiểm tra hợp đồng: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			// Kiểm tra nội dung điều chỉnh
			tenNoiDung, hopLe := noiDungTheoNhom[nhomHopDong][noiDung]

			if !hopLe {
				http.Error(w, "Nội dung điều chỉnh không hợp lệ.",
					http.StatusBadRequest)
				return
			}

			// Sinh mã phụ lục
			maPhuLuc := "PL" + strconv.FormatInt(time.Now().UnixNano(), 10)

			// Lưu vào cơ sở dữ liệu
			_, err = db.Exec(`
        INSERT INTO phuluchopdong (
            ma_phu_luc,
            id_hop_dong,
            nhom_hop_dong,
            noi_dung_dieu_chinh,
            noi_dung_cu,
            noi_dung_moi,
            ngay_ap_dung,
            ly_do
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `,
				maPhuLuc,
				id,
				nhomHopDong,
				tenNoiDung,
				noiDungCu,
				noiDungMoi,
				ngayApDung,
				lyDo,
			)

			if err != nil {
				http.Error(w, "Lỗi lưu phụ lục: "+err.Error(),
					http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/phuluc", http.StatusSeeOther)
			return
		}

		// Lấy thông tin từ bước 1
		idHopDong := strings.TrimSpace(r.URL.Query().Get("id_hop_dong"))
		noiDung := strings.TrimSpace(r.URL.Query().Get("noi_dung_dieu_chinh"))

		id, err := strconv.Atoi(idHopDong)
		if err != nil || id <= 0 {
			http.Error(w, "Mã hợp đồng không hợp lệ.",
				http.StatusBadRequest)
			return
		}

		// Lấy tên nội dung điều chỉnh
		var nhomHopDong string

		err = db.QueryRow(`
			SELECT nhom_hop_dong
			FROM hopdong
			WHERE id = ?
			  AND ngay_bat_dau <= CURDATE()
			  AND (ngay_ket_thuc IS NULL OR ngay_ket_thuc >= CURDATE())
		`, id).Scan(&nhomHopDong)

		if err != nil {
			http.Error(w, "Hợp đồng gốc không tồn tại hoặc đã hết hiệu lực.",
				http.StatusBadRequest)
			return
		}

		tenNoiDung, hopLe := noiDungTheoNhom[nhomHopDong][noiDung]
		if !hopLe {
			http.Error(w, "Nội dung điều chỉnh không hợp lệ.",
				http.StatusBadRequest)
			return
		}

		// Lấy thông tin hợp đồng và nhân viên
		var data struct {
			IDHopDong           int
			MaHD                string
			MaNV                string
			LoaiHopDong         string
			NgayBatDau          string
			NhomHopDong         string
			NoiDungDieuChinh    string
			TenNoiDungDieuChinh string

			HoTen       string
			NgaySinh    string
			GioiTinh    string
			SoDienThoai string
			Email       string
			DiaChi      string
			PhongBan    string
			ChucVu      string
		}

		err = db.QueryRow(`
			SELECT
				hd.id,
				COALESCE(hd.ma_hd, ''),
				COALESCE(hd.ma_nv, ''),
				COALESCE(hd.loai_hop_dong, ''),
				COALESCE(DATE_FORMAT(hd.ngay_bat_dau, '%d/%m/%Y'), ''),
				COALESCE(hd.nhom_hop_dong, ''),
				COALESCE(nv.ho_ten, ''),
				COALESCE(DATE_FORMAT(nv.ngay_sinh, '%d/%m/%Y'), ''),
				COALESCE(nv.gioi_tinh, ''),
				COALESCE(nv.so_dien_thoai, ''),
				COALESCE(nv.email, ''),
				COALESCE(nv.dia_chi, ''),
				COALESCE(nv.phong_ban, ''),
				COALESCE(nv.chuc_vu, '')
			FROM hopdong hd
			LEFT JOIN nhanvien nv ON hd.ma_nv = nv.ma_nv
			WHERE hd.id = ?
			  AND hd.ngay_bat_dau <= CURDATE()
			  AND (hd.ngay_ket_thuc IS NULL OR hd.ngay_ket_thuc >= CURDATE())
		`, id).Scan(
			&data.IDHopDong,
			&data.MaHD,
			&data.MaNV,
			&data.LoaiHopDong,
			&data.NgayBatDau,
			&data.NhomHopDong,
			&data.HoTen,
			&data.NgaySinh,
			&data.GioiTinh,
			&data.SoDienThoai,
			&data.Email,
			&data.DiaChi,
			&data.PhongBan,
			&data.ChucVu,
		)

		if err != nil {
			http.Error(w, "Lỗi lấy thông tin hợp đồng: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		data.NoiDungDieuChinh = noiDung
		data.TenNoiDungDieuChinh = tenNoiDung

		tmpl, err := template.ParseFiles("templates/phuluc/tao.html")
		if err != nil {
			http.Error(w, "Lỗi tải giao diện tạo phụ lục: "+err.Error(),
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
