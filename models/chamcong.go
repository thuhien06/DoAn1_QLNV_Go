package models

import "time"

// ChamCong lưu thông tin chấm công
type ChamCong struct {
	// ID chấm công
	ID int

	// Mã nhân viên
	MaNV string

	// Ngày chấm công
	NgayChamCong *time.Time

	// Giờ vào
	GioVao string

	// Giờ ra
	GioRa string

	// Trạng thái
	TrangThai string

	// Ghi chú
	GhiChu string
}