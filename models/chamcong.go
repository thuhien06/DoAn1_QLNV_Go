package models

import "time"

// ChamCong lưu thông tin chấm công
type ChamCong struct {
	ID           int
	MaNV         string
	NgayChamCong *time.Time
	GioVao       string
	GioRa        string
	TrangThai    string
	GhiChu       string
}
