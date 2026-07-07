package utils

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

func truncate(text string, length int, ellipsis string) (string, int) {
	count := 0
	start := -1
	for end, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if start >= 0 {
				count++
				start = ^start
			}
		} else {
			if start < 0 {
				start = end
			}
		}
		if count >= length {
			return text[:end] + ellipsis, count
		}
	}
	if start >= 0 {
		count++
	}
	return text + ellipsis, count
}

func Truncate(text string, length int, ellipsis string) string {
	t, _ := truncate(text, length, ellipsis)
	return t
}

func TruncateHTML(text string, length int, ellipsis string) string {
	tags := make([]html.Token, 0)
	count := 0

	var (
		b strings.Builder
		z = html.NewTokenizer(strings.NewReader(text))
	)
LOOP:
	for {
		next := z.Next()
		if next == html.ErrorToken {
			break
		}

		raw := string(z.Raw())
		token := z.Token()
		switch next {
		case html.StartTagToken:
			tags = append(tags, token)
		case html.EndTagToken:
			for i := len(tags) - 1; i >= 0; i-- {
				if tags[i].Data == token.Data {
					tags = tags[:i]
					break
				}
				if !IsHTMLVoidElement(tags[i].Data) {
					b.WriteString("</" + tags[i].Data + ">")
				}
			}
		case html.TextToken:
			text, c := truncate(token.Data, length-count, ellipsis)
			count += c

			if count >= length {
				b.WriteString(text)
				break LOOP
			}
		}
		b.WriteString(raw)
	}
	for i := len(tags) - 1; i >= 0; i-- {
		if IsHTMLVoidElement(tags[i].Data) {
			continue
		}
		b.WriteString("</" + tags[i].Data + ">")
	}
	return b.String()
}
