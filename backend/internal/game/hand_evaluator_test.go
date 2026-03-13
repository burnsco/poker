package game

import (
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		cards    []string
		category int
		values   []int
		desc     string
	}{
		{
			name:     "High Card",
			cards:    []string{"As", "Kd", "Qh", "Jc", "9s", "8d", "7h"},
			category: 0,
			values:   []int{14, 13, 12, 11, 9}, // A K Q J 9
			desc:     "High card, Ace",
		},
		{
			name:     "Pair",
			cards:    []string{"As", "Ad", "Qh", "Jc", "9s", "8d", "7h"},
			category: 1,
			values:   []int{14, 12, 11, 9}, // Pair A, kickers Q J 9
			desc:     "Pair of Aces",
		},
		{
			name:     "Two Pair",
			cards:    []string{"As", "Ad", "Ks", "Kd", "Qh", "Jc", "9c"},
			category: 2,
			values:   []int{14, 13, 12}, // Pairs A K, kicker Q
			desc:     "Two pair, Aces and Kings",
		},
		{
			name:     "Three of a Kind",
			cards:    []string{"As", "Ad", "Ah", "Ks", "Qh", "Jc", "9c"},
			category: 3,
			values:   []int{14, 13, 12}, // Trips A, kickers K Q
			desc:     "Three of a kind, Aces",
		},
		{
			name:     "Straight",
			cards:    []string{"5s", "6d", "7h", "8c", "9s", "2d", "3h"},
			category: 4,
			values:   []int{9}, // Straight to 9
			desc:     "Straight, Nine high",
		},
		{
			name:     "Flush",
			cards:    []string{"As", "Qs", "Js", "9s", "7s", "Kd", "Qh"},
			category: 5,
			values:   []int{14, 12, 11, 9, 7}, // Flush A Q J 9 7
			desc:     "Flush, Ace high",
		},
		{
			name:     "Full House",
			cards:    []string{"As", "Ad", "Ah", "Ks", "Kd", "Qh", "Jc"},
			category: 6,
			values:   []int{14, 13}, // Trips A, pair K
			desc:     "Full house, Aces over Kings",
		},
		{
			name:     "Four of a Kind",
			cards:    []string{"As", "Ad", "Ah", "Ac", "Ks", "Qh", "Jc"},
			category: 7,
			values:   []int{14, 13}, // Quads A, kicker K
			desc:     "Four of a kind, Aces",
		},
		{
			name:     "Straight Flush",
			cards:    []string{"5s", "6s", "7s", "8s", "9s", "Kd", "Qh"},
			category: 8,
			values:   []int{9}, // Straight flush to 9
			desc:     "Straight flush, Nine high",
		},
		{
			name:     "Royal Flush",
			cards:    []string{"Ts", "Js", "Qs", "Ks", "As", "2d", "3h"},
			category: 8,
			values:   []int{14}, // Royal flush
			desc:     "Straight flush, Ace high",
		},
		{
			name:     "Ace-low Straight",
			cards:    []string{"As", "2d", "3h", "4c", "5s", "9d", "Kh"},
			category: 4,
			values:   []int{5}, // A-2-3-4-5 straight
			desc:     "Straight, Five high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.cards)
			if res.Score.Category != tt.category {
				t.Errorf("Evaluate() category = %v, want %v", res.Score.Category, tt.category)
			}
			if len(res.Score.Values) != len(tt.values) {
				t.Errorf("Evaluate() values length = %v, want %v", len(res.Score.Values), len(tt.values))
			} else {
				for i, v := range tt.values {
					if res.Score.Values[i] != v {
						t.Errorf("Evaluate() values[%d] = %v, want %v", i, res.Score.Values[i], v)
					}
				}
			}
			if res.Description != tt.desc {
				t.Errorf("Evaluate() desc = %v, want %v", res.Description, tt.desc)
			}
		})
	}
}

func TestScoreCompare(t *testing.T) {
	tests := []struct {
		name     string
		score1   Score
		score2   Score
		expected int // 1 if score1 > score2, -1 if <, 0 if ==
	}{
		{
			name:     "Higher category wins",
			score1:   Score{Category: 1, Values: []int{14}},
			score2:   Score{Category: 0, Values: []int{14, 13}},
			expected: 1,
		},
		{
			name:     "Same category, higher values win",
			score1:   Score{Category: 1, Values: []int{14, 13}},
			score2:   Score{Category: 1, Values: []int{13, 12}},
			expected: 1,
		},
		{
			name:     "Equal scores",
			score1:   Score{Category: 0, Values: []int{14, 13}},
			score2:   Score{Category: 0, Values: []int{14, 13}},
			expected: 0,
		},
		{
			name:     "Two pair comparison",
			score1:   Score{Category: 2, Values: []int{14, 13, 5}}, // AA KK 5
			score2:   Score{Category: 2, Values: []int{14, 12, 6}}, // AA QQ 6
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.score1.Compare(tt.score2)
			if result != tt.expected {
				t.Errorf("Compare() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluateInvalid(t *testing.T) {
	tests := []struct {
		name  string
		cards []string
	}{
		{
			name:  "Too few cards",
			cards: []string{"As", "Kd"},
		},
		{
			name:  "Empty",
			cards: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.cards)
			if len(res.Cards) != 0 {
				t.Errorf("Evaluate() should return invalid for %s", tt.name)
			}
		})
	}
}

func BenchmarkEvaluate(b *testing.B) {
	hands := [][]string{
		{"As", "Kd", "Qh", "Jc", "9s", "8d", "7h"}, // High card
		{"As", "Ad", "Qh", "Jc", "9s", "8d", "7h"}, // Pair
		{"As", "Ad", "Ah", "Ks", "Qh", "Jc", "9c"}, // Trips
		{"As", "Ad", "Ah", "Ac", "Ks", "Qh", "Jc"}, // Quads
		{"Ts", "Js", "Qs", "Ks", "As", "2d", "3h"}, // Royal flush
	}

	for i := 0; i < b.N; i++ {
		for _, hand := range hands {
			Evaluate(hand)
		}
	}
}
