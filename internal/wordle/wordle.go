package wordle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response is the JSON structure from our NYTimes Wordle API response
type Response struct {
	Solution string `json:"solution"`
	Days     int    `json:"days_since_launch"`
	ID       int    `json:"id"`
	Print    string `json:"print_date"`
	Editor   string `json:"editor"`
}

// FetchWord fetches the Wordle word for a given date
func FetchWord(date string) (string, error) {
	url := fmt.Sprintf("https://www.nytimes.com/svc/wordle/v2/%s.json", date)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch wordle data: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("error closing response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch wordle data: status %d", resp.StatusCode)
	}

	var wordleResp Response
	if err := json.NewDecoder(resp.Body).Decode(&wordleResp); err != nil {
		return "", fmt.Errorf("failed to decode wordle data: %w", err)
	}

	return wordleResp.Solution, nil
}

// FetchTodayWord fetches today's Wordle word
func FetchTodayWord() (string, error) {
	today := time.Now().Format("2006-01-02")
	return FetchWord(today)
}
