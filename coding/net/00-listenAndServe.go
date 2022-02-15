package main

import (
	"fmt"
	"log"
	"net/http"
)

// 分析http.ListenAndServe机制
func main() {
	http.HandleFunc("/", indexHandler)
	// 监听TCP地址
	// handler通常设为nil,此时会使用DefaultServeMux(默认路由器)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}
