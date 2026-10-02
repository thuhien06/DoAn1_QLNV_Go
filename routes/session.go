package routes

import "github.com/gorilla/sessions"

//Store sẽ được dùng để tạo và đọc session.
var Store = sessions.NewCookieStore(
	[]byte("qlnv-secret-key-2026"),
)
