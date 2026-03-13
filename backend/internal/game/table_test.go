package game

import (
	"poker-backend/internal/models"
	"testing"
)

func TestConcludeShowdown_SplitPot(t *testing.T) {
	tbl := &Table{
		state: models.TableState{
			Players: []models.Player{
				{
					Seat:                1,
					Name:                "Alice",
					Status:              "ACTIVE",
					HoleCards:           []string{"Ah", "Kh"},
					Stack:               1000,
					ContributedThisHand: 200,
				},
				{
					Seat:                2,
					Name:                "Bob",
					Status:              "ACTIVE",
					HoleCards:           []string{"As", "Ks"},
					Stack:               1000,
					ContributedThisHand: 200,
				},
			},
			HandState: models.HandState{
				Pot:            400,
				CommunityCards: []string{"Ad", "Qc", "Jh", "Th", "2s"}, // Both have Ace-high Straight
				Stage:          "river",
			},
		},
	}

	// Initialize subsystems
	tbl.bettingEngine = NewBettingEngine(&tbl.state, tbl.log)
	tbl.showdownResolver = NewShowdownResolver(&tbl.state, tbl.log)
	tbl.presenceTracker = NewPresenceTracker(&tbl.state, tbl.log)

	tbl.concludeShowdown()

	if len(tbl.state.HandState.WinnerSeats) != 2 {
		t.Errorf("Expected 2 winners, got %d", len(tbl.state.HandState.WinnerSeats))
	}

	expectedAmount := 200

	aliceAmount := tbl.state.HandState.WinnerAmounts["1"]
	bobAmount := tbl.state.HandState.WinnerAmounts["2"]

	if aliceAmount != expectedAmount {
		t.Errorf("Expected Alice to win %d, got %d", expectedAmount, aliceAmount)
	}
	if bobAmount != expectedAmount {
		t.Errorf("Expected Bob to win %d, got %d", expectedAmount, bobAmount)
	}

	if tbl.state.Players[0].Stack != 1200 {
		t.Errorf("Expected Alice stack to be 1200, got %d", tbl.state.Players[0].Stack)
	}
	if tbl.state.Players[1].Stack != 1200 {
		t.Errorf("Expected Bob stack to be 1200, got %d", tbl.state.Players[1].Stack)
	}
}

func TestConcludeShowdown_SplitPot_Remainder(t *testing.T) {
	tbl := &Table{
		state: models.TableState{
			Players: []models.Player{
				{
					Seat:                1,
					Name:                "Alice",
					Status:              "ACTIVE",
					HoleCards:           []string{"Ah", "Kh"},
					Stack:               1000,
					ContributedThisHand: 201,
				},
				{
					Seat:                2,
					Name:                "Bob",
					Status:              "ACTIVE",
					HoleCards:           []string{"As", "Ks"},
					Stack:               1000,
					ContributedThisHand: 200,
				},
			},
			HandState: models.HandState{
				Pot:            401,
				CommunityCards: []string{"Ad", "Qc", "Jh", "Th", "2s"},
				Stage:          "river",
			},
		},
	}

	// Initialize subsystems
	tbl.bettingEngine = NewBettingEngine(&tbl.state, tbl.log)
	tbl.showdownResolver = NewShowdownResolver(&tbl.state, tbl.log)
	tbl.presenceTracker = NewPresenceTracker(&tbl.state, tbl.log)

	tbl.concludeShowdown()

	aliceAmount := tbl.state.HandState.WinnerAmounts["1"]
	bobAmount := tbl.state.HandState.WinnerAmounts["2"]

	if aliceAmount+bobAmount != 401 {
		t.Errorf("Expected total win to be 401, got %d", aliceAmount+bobAmount)
	}

	if (aliceAmount == 201 && bobAmount == 200) || (aliceAmount == 200 && bobAmount == 201) {
		// OK
	} else {
		t.Errorf("Expected split to be 201/200, got %d/%d", aliceAmount, bobAmount)
	}
}

func TestConcludeShowdown_Foldout(t *testing.T) {
	tbl := &Table{
		state: models.TableState{
			Players: []models.Player{
				{
					Seat:                1,
					Name:                "Alice",
					Status:              "FOLDED",
					Stack:               1000,
					ContributedThisHand: 0,
				},
				{
					Seat:                2,
					Name:                "Bob",
					Status:              "ACTIVE",
					Stack:               1000,
					ContributedThisHand: 200,
				},
			},
			HandState: models.HandState{
				Pot:            200,
				CommunityCards: []string{"Ad", "Qc", "Jh", "Th", "2s"},
				Stage:          "river",
			},
		},
	}

	// Initialize subsystems
	tbl.bettingEngine = NewBettingEngine(&tbl.state, tbl.log)
	tbl.showdownResolver = NewShowdownResolver(&tbl.state, tbl.log)
	tbl.presenceTracker = NewPresenceTracker(&tbl.state, tbl.log)

	tbl.concludeShowdown()

	if len(tbl.state.HandState.WinnerSeats) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(tbl.state.HandState.WinnerSeats))
	}

	bobWin := tbl.state.HandState.WinnerAmounts["2"]
	if bobWin != 200 {
		t.Errorf("Expected Bob to win 200, got %d", bobWin)
	}

	if tbl.state.Players[1].Stack != 1200 {
		t.Errorf("Expected Bob stack to be 1200, got %d", tbl.state.Players[1].Stack)
	}
}

func TestLeave_DisconnectsPlayer(t *testing.T) {
	playerID := "test@example.com"
	tbl := &Table{
		state: models.TableState{
			Players: []models.Player{
				{
					Seat:      1,
					Name:      "Test Player",
					PlayerID:  &playerID,
					IsBot:     false,
					Connected: true,
				},
			},
			ClientConnections: map[string]int{playerID: 1},
		},
	}

	// Initialize subsystems
	tbl.bettingEngine = NewBettingEngine(&tbl.state, tbl.log)
	tbl.showdownResolver = NewShowdownResolver(&tbl.state, tbl.log)
	tbl.presenceTracker = NewPresenceTracker(&tbl.state, tbl.log)

	tbl.Leave(playerID)

	if tbl.state.Players[0].Connected {
		t.Errorf("Expected player to be disconnected")
	}
	if tbl.state.ClientConnections[playerID] != 0 {
		t.Errorf("Expected client connections to be 0, got %d", tbl.state.ClientConnections[playerID])
	}
}

func TestTable_NextHand(t *testing.T) {
	tbl := NewTable("test-table", false)
	
	// Add 3 players
	p1 := "p1"
	p2 := "p2"
	p3 := "p3"
	tbl.Join(p1, "Alice")
	tbl.Join(p2, "Bob")
	tbl.Join(p3, "Charlie")

	// Sit them in
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": p1, "seat": 1, "player_name": "Alice"})
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": p2, "seat": 2, "player_name": "Bob"})
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": p3, "seat": 3, "player_name": "Charlie"})

	// Set initial dealer to seat 3
	tbl.state.HandState.DealerSeat = 3
	
	// Start next hand
	// Operator check: player_id "1" is operator by default in Table.isOperator
	tbl.ApplyAction("next_hand", map[string]interface{}{"player_id": "1"})

	// Dealer rotation: 3 -> 1
	if tbl.state.HandState.DealerSeat != 1 {
		t.Errorf("Expected dealer seat 1, got %d", tbl.state.HandState.DealerSeat)
	}
	// SB should be 2, BB should be 3
	if tbl.state.HandState.SmallBlindSeat != 2 {
		t.Errorf("Expected SB seat 2, got %d", tbl.state.HandState.SmallBlindSeat)
	}
	if tbl.state.HandState.BigBlindSeat != 3 {
		t.Errorf("Expected BB seat 3, got %d", tbl.state.HandState.BigBlindSeat)
	}
	// Action on seat 1 (because heads-up is special, but for 3 players, action starts after BB)
	// For 3 players: Dealer(1), SB(2), BB(3). Action on 1? 
	// Wait, nextHandPositions: 
	// dealer := nextSeatInList(readySeats, 3) -> 1
	// sb := nextSeatInList(readySeats, 1) -> 2
	// bb := nextSeatInList(readySeats, 2) -> 3
	// acting := nextSeatInList(readySeats, 3) -> 1
	if *tbl.state.HandState.ActingSeat != 1 {
		t.Errorf("Expected acting seat 1, got %d", *tbl.state.HandState.ActingSeat)
	}

	// Verify blinds were posted
	if tbl.state.Players[1].BetThisStreet != SmallBlind { // Seat 2 is index 1
		t.Errorf("Expected seat 2 to post SB %d, got %d", SmallBlind, tbl.state.Players[1].BetThisStreet)
	}
	if tbl.state.Players[2].BetThisStreet != BigBlind { // Seat 3 is index 2
		t.Errorf("Expected seat 3 to post BB %d, got %d", BigBlind, tbl.state.Players[2].BetThisStreet)
	}
}

func TestTable_HandInitialization(t *testing.T) {
	tbl := NewTable("test-table", false)
	tbl.Join("p1", "Alice")
	tbl.Join("p2", "Bob")
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": "p1", "seat": 1, "player_name": "Alice"})
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": "p2", "seat": 2, "player_name": "Bob"})

	tbl.ApplyAction("next_hand", map[string]interface{}{"player_id": "1"})

	if tbl.state.HandState.Status != "in_progress" {
		t.Errorf("Expected hand status in_progress, got %s", tbl.state.HandState.Status)
	}
	if tbl.state.HandState.Stage != "preflop" {
		t.Errorf("Expected stage preflop, got %s", tbl.state.HandState.Stage)
	}
	if len(tbl.state.HandState.Deck) != 52-4 { // 2 players * 2 cards
		t.Errorf("Expected deck to have 48 cards, got %d", len(tbl.state.HandState.Deck))
	}
	for i := 0; i < 2; i++ {
		if tbl.state.Players[i].HoleCards[0] == "" || tbl.state.Players[i].HoleCards[1] == "" {
			t.Errorf("Expected player %d to have hole cards", i+1)
		}
	}
}
