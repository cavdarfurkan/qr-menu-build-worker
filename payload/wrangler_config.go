package payload

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	base_site = "menu.furkancavdar.com"
)

type routesType struct {
	Pattern      string `json:"pattern"`
	CustomDomain bool   `json:"custom_domain"`
}

type assetsType struct {
	Directory        string `json:"directory"`
	NotFoundHandling string `json:"not_found_handling"`
}

type WranglerConfig struct {
	Name              string       `json:"name"`
	Routes            []routesType `json:"routes"`
	CompatibilityDate string       `json:"compatibility_date"`
	Assets            assetsType   `json:"assets"`
}

func NewWranglerConfig(siteName string) WranglerConfig {
	routes := []routesType{
		{
			Pattern:      fmt.Sprintf("%s.%s", siteName, base_site),
			CustomDomain: true,
		},
	}

	assets := assetsType{
		Directory:        "./dist",
		NotFoundHandling: "404-page",
	}

	return WranglerConfig{
		Name:              siteName,
		Routes:            routes,
		CompatibilityDate: time.Now().UTC().Format("2006-01-02"),
		Assets:            assets,
	}
}

func (wc *WranglerConfig) MarshalConfig() (string, error) {
	jsonBytes, err := json.Marshal(wc)
	if err != nil {
		return "", fmt.Errorf("Error marshaling to JSON: %w", err)
	}
	return string(jsonBytes), nil
}
