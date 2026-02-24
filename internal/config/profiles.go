package config

import (
	"sort"
	"strings"
)

func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Config) SuggestProfileNames(target string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		return nil
	}

	target = strings.TrimSpace(target)
	if target == "" {
		return nil
	}

	targetLower := strings.ToLower(target)
	type suggestion struct {
		name     string
		score    int
		distance int
	}

	allNames := c.ProfileNames()
	ranked := make([]suggestion, 0, len(allNames))
	for _, name := range allNames {
		nameLower := strings.ToLower(name)
		distance := levenshtein(nameLower, targetLower)

		var score int
		switch {
		case nameLower == targetLower:
			score = 0
		case strings.HasPrefix(nameLower, targetLower) || strings.HasPrefix(targetLower, nameLower):
			score = 1
		case strings.Contains(nameLower, targetLower) || strings.Contains(targetLower, nameLower):
			score = 2
		case distance <= 2:
			score = 3
		case distance <= 3:
			score = 4
		default:
			continue
		}

		ranked = append(ranked, suggestion{name: name, score: score, distance: distance})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score < ranked[j].score
		}
		if ranked[i].distance != ranked[j].distance {
			return ranked[i].distance < ranked[j].distance
		}
		return ranked[i].name < ranked[j].name
	})

	if len(ranked) > maxSuggestions {
		ranked = ranked[:maxSuggestions]
	}

	result := make([]string, 0, len(ranked))
	for _, item := range ranked {
		result = append(result, item.name)
	}
	return result
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return len(b)
	}
	if b == "" {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			deletion := prev[j] + 1
			insertion := current[j-1] + 1
			substitution := prev[j-1] + cost

			current[j] = minInt(deletion, insertion, substitution)
		}
		prev = current
	}

	return prev[len(b)]
}

func minInt(values ...int) int {
	minVal := values[0]
	for _, value := range values[1:] {
		if value < minVal {
			minVal = value
		}
	}
	return minVal
}
