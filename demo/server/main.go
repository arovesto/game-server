package main

import (
	"net/http"
)


func main() {
	if err := http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`static`))); err != http.ErrServerClosed {
		panic(err)
	}
}
