package api

import "net/http"

func CookieButton(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./cookie_button.html")
}
