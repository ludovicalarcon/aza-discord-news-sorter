package helpers

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func GetTitleFromUrl(url string) (title string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "title" {
			if z.Next() == html.TextToken {
				title = strings.TrimSpace(z.Token().Data)
				return
			}
		}

	}
	return
}
