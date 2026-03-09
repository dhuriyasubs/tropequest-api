package services

import "strings"

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			dp[i][j] = min3(
				dp[i-1][j]+1,
				dp[i][j-1]+1,
				dp[i-1][j-1]+cost,
			)
		}
	}
	return dp[la][lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// fuzzyTextScore returns a score [0,1] for how well queryWords match
// against a target text using exact + fuzzy word matching.
func FuzzyTextScore(query, target string) float64 {
	if query == "" || target == "" {
		return 0
	}

	queryWords := strings.Fields(strings.ToLower(query))
	targetWords := strings.Fields(strings.ToLower(target))

	if len(queryWords) == 0 || len(targetWords) == 0 {
		return 0
	}

	totalScore := 0.0
	for _, qw := range queryWords {
		if len(qw) < 2 {
			continue
		}
		best := 0.0
		for _, tw := range targetWords {
			if len(tw) < 2 {
				continue
			}
			// Exact match
			if qw == tw {
				best = 1.0
				break
			}
			// Prefix match (e.g. "coll" matches "colleen")
			if strings.HasPrefix(tw, qw) || strings.HasPrefix(qw, tw) {
				s := 0.85
				if s > best {
					best = s
				}
				continue
			}
			// Fuzzy: allow up to 2 edits for words ≥ 4 chars
			maxLen := len(qw)
			if len(tw) > maxLen {
				maxLen = len(tw)
			}
			if maxLen < 4 {
				continue
			}
			dist := levenshtein(qw, tw)
			threshold := 1
			if maxLen >= 6 {
				threshold = 2
			}
			if dist <= threshold {
				s := 1.0 - float64(dist)/float64(maxLen)
				if s > best {
					best = s
				}
			}
		}
		totalScore += best
	}

	return totalScore / float64(len(queryWords))
}
