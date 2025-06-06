package main

import (
	"errors"
	"fmt"
	u "net/url"
	"strings"

	"golang.org/x/net/html"
)

func NormalizeURL(url string) string {
	var parsedURL *u.URL
	parsedURL, err := u.Parse(url)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s%s", parsedURL.Host, parsedURL.Path)
}

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	urls := []string{}
	buff := strings.NewReader(htmlBody)
	z := html.NewTokenizer(buff)
	for {
		token := z.Next()
		if token == html.ErrorToken {
			return urls, z.Err()

		}
		if token == html.StartTagToken {
			tokenData := z.Token()
			if tokenData.Data == "a" {
				for _, attr := range tokenData.Attr {
					if attr.Key == "href" {
						url, err := u.Parse(attr.Val)
						if err != nil {
							return nil, errors.New("error parsing URL")
						}
						if url.IsAbs() {
							urls = append(urls, url.String())
						} else {
							baseURL, err := u.Parse(rawBaseURL)
							if err != nil {
								return nil, errors.New("error parsing base URL")
							}
							urls = append(urls, baseURL.ResolveReference(url).String())
						}
					}
				}
			}
		}
		if token == html.EndTagToken {
			tokenData := z.Token()
			if tokenData.Data == "a" {
				continue
			}
		}
	}
}
