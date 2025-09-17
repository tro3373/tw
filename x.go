package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	twscraper "github.com/imperatrona/twitter-scraper"
)

type XClient struct {
	Scraper *twscraper.Scraper
}

func NewXClient() (client *XClient, err error) {
	scraper := twscraper.New()
	cookieJSONPath := os.Getenv("X_COOKIE_JSON")
	if cookieJSONPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		cookieJSONPath = filepath.Join(home, ".cookies.json")
	}
	if err := setupCookieAuth(scraper, cookieJSONPath); err != nil {
		return nil, err
	}
	if !scraper.IsLoggedIn() {
		return nil, fmt.Errorf("invalid credentials: failed to authenticate with X/Twitter")
	}
	return &XClient{
		Scraper: scraper,
	}, nil
}

func setupCookieAuth(scraper *twscraper.Scraper, cookieJSONPath string) (err error) {

	var cookiesJSON []CookieJSON
	// if !filepath.IsAbs(cookieJSONPath) || strings.Contains(cookieJSONPath, "..") {
	// 	return fmt.Errorf("invalid cookie JSON path: %s. Please provide an absolute path or a valid relative path without '..'", cookieJSONPath)
	// }
	f, err := os.Open(filepath.Clean(cookieJSONPath))
	if err != nil {
		return fmt.Errorf("failed to open config file %s: %w", cookieJSONPath, err)
	}

	defer func() {
		cerr := f.Close()
		if cerr == nil {
			return
		}
		if err == nil {
			err = fmt.Errorf("failed to close cookies.json: %w", cerr)
			return
		}
		err = fmt.Errorf("%v; additionally failed to close cookies.json: %w", err, cerr)
	}()

	err = json.NewDecoder(f).Decode(&cookiesJSON)
	if err != nil {
		return fmt.Errorf("failed to decode cookies: %w", err)
	}

	var cookies []*http.Cookie
	for _, c := range cookiesJSON {
		httpCookie := c.ToHTTPCookie()
		cookies = append(cookies, httpCookie)
	}

	for _, cookie := range cookies {
		switch cookie.Domain {
		case ".x.com":
			cookie.Domain = ".twitter.com"
		case "x.com":
			cookie.Domain = "twitter.com"
		}
	}
	scraper.SetCookies(cookies)
	return nil
}

// CookieJSON represents a cookie in JSON format (browser export format)
type CookieJSON struct {
	Name           string  `json:"name"`
	Value          string  `json:"value"`
	Path           string  `json:"path"`
	Domain         string  `json:"domain"`
	Secure         bool    `json:"secure"`
	HTTPOnly       bool    `json:"httpOnly"`
	SameSite       string  `json:"sameSite"`
	ExpirationDate float64 `json:"expirationDate,omitempty"`
	HostOnly       bool    `json:"hostOnly,omitempty"`
	Session        bool    `json:"session,omitempty"`
	StoreID        string  `json:"storeId,omitempty"`
	ID             int     `json:"id,omitempty"`
}

// ToHTTPCookie converts CookieJSON to http.Cookie
func (c *CookieJSON) ToHTTPCookie() *http.Cookie {
	// Remove surrounding quotes from value if present
	value := c.Value
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	cookie := &http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     c.Path,
		Domain:   c.Domain,
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
	}

	// Convert string SameSite to http.SameSite
	switch strings.ToLower(c.SameSite) {
	case "lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "none", "no_restriction":
		cookie.SameSite = http.SameSiteNoneMode
	default:
		cookie.SameSite = http.SameSiteDefaultMode
	}

	return cookie
}
