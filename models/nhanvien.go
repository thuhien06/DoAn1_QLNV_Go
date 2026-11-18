package models

import "time"

type NhanVien struct {
	ID          int
	MaNV        string
	HoTen       string
	NgaySinh    *time.Time
	GioiTinh    string
	SoDienThoai string
	Email       string
	DiaChi      string
	NgayVaoLam  *time.Time
	PhongBan    string
	ChucVu      string

	QuocTich        string
	NoiSinh         string
	CCCD            string
	NgayCapCCCD     *time.Time
	NoiCapCCCD      string
	DiaChiThuongTru string
	DiaChiTamTru    string
}
