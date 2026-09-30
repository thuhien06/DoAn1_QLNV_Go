package models

import "time"

// NghiPhep lưu thông tin nghỉ phép của nhân viên
type NghiPhep struct {
	ID           int
	MaNV         string
	NgayBatDau   *time.Time
	NgayKetThuc  *time.Time
	LoaiNghi     string
	LyDo         string
	TrangThai    string
	GhiChu       string
}