package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const githubReleasesAPI = "https://api.github.com/repos/BlackMetalz/lazy-hole/releases/latest"

// updateState represents the result of the version check
type updateState int

const (
	updateStateChecking  updateState = iota // still in flight
	updateStateUpToDate                     // current == latest
	updateStateAvailable                    // newer version on github
	updateStateUnavail                      // could not reach github
)

// UpdateInfo holds the result of a version check.
type UpdateInfo struct {
	State     updateState
	LatestTag string // populated when state == updateStateAvailable
}

// githubRelease is the minimal shape we need from the API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckUpdateAsync fires a goroutine that hits the GitHub releases API once
// and writes the result into the returned channel. The channel is closed after
// one send so callers can safely range over it or do a single receive.
func CheckUpdateAsync(currentVersion string) <-chan UpdateInfo {
	ch := make(chan UpdateInfo, 1)

	go func() {
		defer close(ch)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(githubReleasesAPI)
		if err != nil {
			ch <- UpdateInfo{State: updateStateUnavail}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- UpdateInfo{State: updateStateUnavail}
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ch <- UpdateInfo{State: updateStateUnavail}
			return
		}

		var release githubRelease
		if err := json.Unmarshal(body, &release); err != nil || release.TagName == "" {
			ch <- UpdateInfo{State: updateStateUnavail}
			return
		}

		latest := strings.TrimSpace(release.TagName)
		current := strings.TrimSpace(currentVersion)

		if latest == current {
			ch <- UpdateInfo{State: updateStateUpToDate, LatestTag: latest}
		} else {
			ch <- UpdateInfo{State: updateStateAvailable, LatestTag: latest}
		}
	}()

	return ch
}

// UpdateBannerText returns the coloured tview string to embed in the header.
func UpdateBannerText(info UpdateInfo) string {
	switch info.State {
	case updateStateChecking:
		return "[gray]Checking for updates...[-]"
	case updateStateUpToDate:
		return fmt.Sprintf("[green]✓ Up to date (%s)[-]", info.LatestTag)
	case updateStateAvailable:
		return fmt.Sprintf("[yellow]⬆ Update available: %s[-]", info.LatestTag)
	case updateStateUnavail:
		return "[gray]Update check unavailable[-]"
	}
	return ""
}
