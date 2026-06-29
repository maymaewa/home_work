package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

const topSize = 10

type wordStat struct {
	word  string
	count int
}

func Top10(text string) []string {
	freq := getFreqMap(text)
	stats := getWordStat(freq)
	sortWordStat(stats)
	return getTop(stats)
}

func getFreqMap(text string) map[string]int {
	words := strings.Fields(text)

	freq := make(map[string]int)
	for _, word := range words {
		freq[word]++
	}

	return freq
}

func getWordStat(freq map[string]int) []wordStat {
	stats := make([]wordStat, 0, len(freq))
	for word, count := range freq {
		stats = append(stats, wordStat{
			word:  word,
			count: count,
		})
	}
	return stats
}

func sortWordStat(stats []wordStat) {
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].count == stats[j].count {
			return stats[i].word < stats[j].word
		}

		return stats[i].count > stats[j].count
	})
}

func getTop(stats []wordStat) []string {
	limit := min(topSize, len(stats))

	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, stats[i].word)
	}

	return result
}
