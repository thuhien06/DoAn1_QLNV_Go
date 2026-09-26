package main

import (
	"fmt"
	"net/http"

	"qlnv/database"
	"qlnv/routes"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "QLNV Go dang chay!")
}

func main() {

	db, err := database.Connect()

	if err != nil {
		fmt.Println("Loi ket noi MySQL:", err)
		return
	}

	defer db.Close()

	http.HandleFunc("/", home)

	http.HandleFunc("/nhanvien", routes.DanhSachNhanVien(db))

	fmt.Println("Server dang chay tai http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Loi:", err)
	}
}
