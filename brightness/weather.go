package brightness

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// openMeteoResponse is the subset of the Open-Meteo API response we need.
type openMeteoResponse struct {
	Current struct {
		CloudCover int `json:"cloud_cover"`
	} `json:"current"`
}

// FetchCloudCover returns the current cloud cover percentage (0-100).
func FetchCloudCover(client *http.Client, baseURL string, lat, lon float64) (int, error) {
	url := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f&current=cloud_cover",
		baseURL, lat, lon,
	)

	resp, err := client.Get(url) //nolint:noctx // simple fire-and-forget HTTP call
	if err != nil {
		return 0, fmt.Errorf("weather request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weather: status %d", resp.StatusCode)
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("weather decode: %w", err)
	}
	return data.Current.CloudCover, nil
}
