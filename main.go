package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		println("no website provided\r\nusage: ./go run main.go <URL>")
		os.Exit(1)
	}
	if len(args) > 1 {
		println("too many arguments provided, only one URL is expected")
		os.Exit(1)
	}
	println("starting crawl of: ", args[0])
	htmlBody, err := getHTML(args[0])
	if err != nil {
		println("error fetching URL:", err.Error())
		os.Exit(1)
	}
	print(htmlBody)

}
func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		return "", errors.New(fmt.Sprintf("expected Content-Type 'text/html', got '%s'", resp.Header.Get("Content-Type")))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
