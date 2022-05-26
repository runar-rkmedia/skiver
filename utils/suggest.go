package utils

import (
	"sort"
	"strings"
)

type Suggestion struct {
	Index               int
	Value               string
	LevenshteinDistance int
	From                string
}

type Suggestions []Suggestion

func (s Suggestions) Len() int {
	return len(s)
}
func (s Suggestions) Less(i, j int) bool {
	// TODO: how do we sort when there is a match on prefix?
	diff := s[i].LevenshteinDistance - s[j].LevenshteinDistance
	if diff != 0 {
		return diff < 0
	}
	return len(s[i].Value) < len(s[j].Value)
}
func (s Suggestions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Suggestions) ToStringSlice() []string {
	var slice []string
	for _, sug := range s {
		slice = append(slice, sug.Value)
	}
	return slice
}

func SuggestionsFor(typedName string, stringSlice []string, maxDistance int, maxItems int) Suggestions {
	if maxDistance == 0 {
		maxDistance = 3
	}
	suggestions := Suggestions{}
	for i, s := range stringSlice {
		levenshteinDistance := ld(typedName, s, true)
		suggestByLevenshtein := levenshteinDistance <= maxDistance
		if suggestByLevenshtein {
			suggestions = append(suggestions, Suggestion{
				Index:               i,
				Value:               s,
				LevenshteinDistance: levenshteinDistance,
				From:                "Levenshtein",
			})
			continue
		}
		suggestByPrefix := strings.HasPrefix(strings.ToLower(s), strings.ToLower(typedName))
		if suggestByPrefix {
			suggestions = append(suggestions, Suggestion{
				Index:               i,
				Value:               s,
				LevenshteinDistance: levenshteinDistance,
				From:                "prefixEquality",
			})
			continue
		}
		if strings.EqualFold(typedName, s) {
			suggestions = append(suggestions, Suggestion{
				Index:               i,
				Value:               s,
				LevenshteinDistance: levenshteinDistance,
				From:                "Fold",
			})
			continue
		}
	}
	sort.Sort(suggestions)

	if maxItems > 0 && len(suggestions) >= maxItems {
		return suggestions[0:maxItems]
	}
	return suggestions
}

// levenshteinDistance
func ld(s, t string, ignoreCase bool) int {
	if ignoreCase {
		s = strings.ToLower(s)
		t = strings.ToLower(t)
	}
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}
	return d[len(s)][len(t)]
}
