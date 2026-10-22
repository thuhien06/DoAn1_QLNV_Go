package routes

import (
	"html/template"
	"net/http"
)

func QuanLyCong() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(
			template.ParseFiles("templates/cong/cong.html"),
		)

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func CongCaNhan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(
			template.ParseFiles("templates/cong/canhan.html"),
		)

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
