package main

import "net/http"

func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:9999/", nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

}
