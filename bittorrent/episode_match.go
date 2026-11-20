package bittorrent

import (
	"regexp"
	"strconv"
	"strings"
)

// Digits only: the folder may say Season, Сезон, Staffel or nothing at all.
// 01 is season 1, 11 is season 11.
var (
	numberPattern        = regexp.MustCompile(`\d+`)
	leadingNumberPattern = regexp.MustCompile(`^(\d+)`)
)

// matchSeasonFolderEpisode reports whether a parent directory contains the
// season number and the file name starts with the episode number.
// The nearest directory that has a number decides, so Season_02 is not season 1
// even if a parent folder also contains 1.
func matchSeasonFolderEpisode(path string, season, episode int) bool {
	if path == "" || season < 0 || episode < 0 {
		return false
	}

	parts := strings.Split(strings.ReplaceAll(path, "\\", "/"), "/")
	if len(parts) < 2 || !fileStartsWithEpisode(parts[len(parts)-1], episode) {
		return false
	}

	for i := len(parts) - 2; i >= 0; i-- {
		dir := parts[i]
		if dir == "" || dir == "." {
			continue
		}
		numbers := numberPattern.FindAllString(dir, -1)
		if len(numbers) == 0 {
			continue
		}
		for _, raw := range numbers {
			number, err := strconv.Atoi(raw)
			if err == nil && number == season {
				return true
			}
		}
		return false
	}

	return false
}

func countSeasonFolderMatches(choices []*CandidateFile, season, episode int) (index, found int) {
	index = -1
	for i, choice := range choices {
		if choice != nil && matchSeasonFolderEpisode(choice.Path, season, episode) {
			index = i
			found++
		}
	}
	return
}

func fileStartsWithEpisode(name string, episode int) bool {
	match := leadingNumberPattern.FindStringSubmatch(strings.TrimSpace(name))
	if match == nil {
		return false
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return false
	}
	return number == episode
}
