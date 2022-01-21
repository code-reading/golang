package main

import (
	"fmt"
	"log"
	"net/http"
)

// 分析http.ListenAndServe机制
func main() {
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}
