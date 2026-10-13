package models

import "time"

// HopDong lưu thông tin các nhóm hợp đồng
type HopDong struct {
	ID          int
	MaHD        string
	MaNV        string
	NhomHopDong string
	LoaiHopDong string
	NgayBatDau  *time.Time
	NgayKetThuc *time.Time
	LuongCoBan  float64
	GhiChu      string

	DoiTac        *string
	NguoiDaiDien  *string
	NoiDung       *string
	GiaTri        *float64
	NgayKy        *time.Time
	DichVuThietBi *string
	ChiPhi        *float64

	TrangThai string
}
