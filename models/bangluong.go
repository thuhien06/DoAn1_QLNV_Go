package models

// BangLuong lưu thông tin bảng lương của nhân viên
type BangLuong struct {
	ID         int
	MaNV       string
	Thang      int
	Nam        int
	LuongCoBan float64
	SoNgayCong int
	PhuCap     float64
	KhauTru    float64
	TongLuong  float64
	GhiChu     string
}
