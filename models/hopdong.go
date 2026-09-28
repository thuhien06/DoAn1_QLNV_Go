package models

import "time"

// HopDong lưu thông tin hợp đồng lao động
type HopDong struct {
	// ID của hợp đồng
	ID int

	// Mã hợp đồng
	MaHD string

	// Mã nhân viên
	MaNV string

	// Loại hợp đồng
	LoaiHopDong string

	// Ngày bắt đầu
	NgayBatDau *time.Time

	// Ngày kết thúc
	NgayKetThuc *time.Time

	// Lương cơ bản
	LuongCoBan float64

	// Ghi chú
	GhiChu string
}
