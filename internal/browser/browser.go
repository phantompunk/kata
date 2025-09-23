package browser

import (
	"context"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
)

// GetCookies retrieves cookies for the domain "leetcode.com" from installed browsers
func GetCookies() (map[string]string, error) {
	cookies := make(map[string]string)

	rows := kooky.TraverseCookies(context.Background(), kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`)).OnlyCookies()
	for row := range rows {
		cookies[row.Name] = row.Value
	}
	return cookies, nil
}
