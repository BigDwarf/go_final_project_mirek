package server

import "net/http"

func MainHandle (res http.ResponseWriter, req *http.Request) {
	out := "Hello from server package!"
	res.Write([]byte(out))
}