package game

import (
	"poker-backend/internal/models"
	"testing"
)

func TestBettingEngine_ProcessAction_Fold(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat: &[]int{1}[0], // Seat 1
			CurrentBet: 20,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 10, Stack: 1000},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 0, Stack: 1000},
			{Seat: 3, Status: "ACTIVE", BetThisStreet: 0, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	showdown, err := be.ProcessAction("fold", "player1", map[string]interface{}{}, func(int, string) bool { return true })

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if showdown {
		t.Errorf("Fold should not trigger showdown")
	}
	if state.Players[0].Status != "FOLDED" {
		t.Errorf("Player should be folded")
	}
}

func TestBettingEngine_ProcessAction_Check(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat: &[]int{1}[0],
			CurrentBet: 10,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 10, Stack: 1000},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 10, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	showdown, err := be.ProcessAction("check", "player1", map[string]interface{}{}, func(int, string) bool { return true })

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if showdown {
		t.Errorf("Check should not trigger showdown")
	}
}

func TestBettingEngine_ProcessAction_Call(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat: &[]int{1}[0],
			CurrentBet: 20,
			Pot:        10,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 10, Stack: 1000, ContributedThisHand: 10},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 20, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	showdown, err := be.ProcessAction("call", "player1", map[string]interface{}{}, func(int, string) bool { return true })

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if showdown {
		t.Errorf("Call should not trigger showdown")
	}
	if state.Players[0].BetThisStreet != 20 {
		t.Errorf("BetThisStreet should be 20, got %d", state.Players[0].BetThisStreet)
	}
	if state.Players[0].Stack != 990 {
		t.Errorf("Stack should be 990, got %d", state.Players[0].Stack)
	}
	if state.HandState.Pot != 20 {
		t.Errorf("Pot should be 20, got %d", state.HandState.Pot)
	}
}

func TestBettingEngine_ProcessAction_Bet(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat:   &[]int{1}[0],
			CurrentBet:   0,
			MinimumRaise: 20,
			Pot:          0,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 0, Stack: 1000, ContributedThisHand: 0},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 0, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	showdown, err := be.ProcessAction("bet", "player1", map[string]interface{}{"amount": 50}, func(int, string) bool { return true })

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if showdown {
		t.Errorf("Bet should not trigger showdown")
	}
	if state.HandState.CurrentBet != 50 {
		t.Errorf("CurrentBet should be 50, got %d", state.HandState.CurrentBet)
	}
	if state.Players[0].BetThisStreet != 50 {
		t.Errorf("BetThisStreet should be 50, got %d", state.Players[0].BetThisStreet)
	}
	if state.Players[0].Stack != 950 {
		t.Errorf("Stack should be 950, got %d", state.Players[0].Stack)
	}
}

func TestBettingEngine_ProcessAction_Raise(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat:   &[]int{1}[0],
			CurrentBet:   20,
			MinimumRaise: 20,
			Pot:          20,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 20, Stack: 1000, ContributedThisHand: 20},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 20, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	showdown, err := be.ProcessAction("raise", "player1", map[string]interface{}{"amount": 60}, func(int, string) bool { return true })

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if showdown {
		t.Errorf("Raise should not trigger showdown")
	}
	if state.HandState.CurrentBet != 60 {
		t.Errorf("CurrentBet should be 60, got %d", state.HandState.CurrentBet)
	}
	if state.HandState.MinimumRaise != 40 {
		t.Errorf("MinimumRaise should be 40, got %d", state.HandState.MinimumRaise)
	}
}

func TestBettingEngine_ProcessAction_InvalidActor(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat: &[]int{1}[0],
		},
		Players: []models.Player{
			{Seat: 1, Status: "FOLDED"},
			{Seat: 2, Status: "ACTIVE"},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	_, err := be.ProcessAction("check", "player1", map[string]interface{}{}, func(int, string) bool { return true })

	if err == nil {
		t.Errorf("Expected error for invalid actor")
	}
}

func TestBettingEngine_ProcessAction_Unauthorized(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat: &[]int{1}[0],
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE"},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	_, err := be.ProcessAction("check", "player1", map[string]interface{}{}, func(int, string) bool { return false })

	if err == nil {
		t.Errorf("Expected unauthorized error")
	}
}

func TestBettingEngine_ProcessAction_IllegalBet(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat:   &[]int{1}[0],
			CurrentBet:   20,
			MinimumRaise: 20,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 20, Stack: 1000, ContributedThisHand: 20},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 20, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	// Bet less than minimum raise
	_, err := be.ProcessAction("raise", "player1", map[string]interface{}{"amount": 30}, func(int, string) bool { return true })

	if err == nil {
		t.Errorf("Expected error for illegal raise amount")
	}
}

func TestBettingEngine_ProcessAction_AllInBet(t *testing.T) {
	state := &models.TableState{
		HandState: models.HandState{
			ActingSeat:   &[]int{1}[0],
			CurrentBet:   0,
			MinimumRaise: 20,
			Pot:          0,
		},
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", BetThisStreet: 0, Stack: 100, ContributedThisHand: 0},
			{Seat: 2, Status: "ACTIVE", BetThisStreet: 0, Stack: 1000},
		},
	}
	log := func(string) {}
	be := NewBettingEngine(state, log)

	showdown, err := be.ProcessAction("bet", "player1", map[string]any{"amount": 100}, func(int, string) bool { return true })

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if showdown {
		t.Errorf("Bet should not trigger showdown")
	}
	if state.Players[0].Status != "ALL_IN" {
		t.Errorf("Player should be ALL_IN")
	}
	if state.Players[0].Stack != 0 {
		t.Errorf("Stack should be 0, got %d", state.Players[0].Stack)
	}
	if state.HandState.CurrentBet != 100 {
		t.Errorf("CurrentBet should be 100, got %d", state.HandState.CurrentBet)
	}
}

func TestBettingEngine_MultiStreet(t *testing.T) {
	state := &models.TableState{
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", Stack: 1000, Name: "Alice", PlayerID: &[]string{"p1"}[0]},
			{Seat: 2, Status: "ACTIVE", Stack: 1000, Name: "Bob", PlayerID: &[]string{"p2"}[0]},
		},
		HandState: models.HandState{
			Status:     "in_progress",
			Stage:      "preflop",
			CurrentBet: 20,
			ActingSeat: &[]int{1}[0],
			DealerSeat: 2,
			Deck:       []string{"As", "Ks", "Qs", "Js", "Ts"}, // Cards for flop, turn, river
		},
	}
	log := func(msg string) { t.Log(msg) }
	be := NewBettingEngine(state, log)
	auth := func(seat int, pid string) bool {
		if seat == 1 && pid == "p1" {
			return true
		}
		if seat == 2 && pid == "p2" {
			return true
		}
		return false
	}

	// Preflop: Seat 1 calls
	_, err := be.ProcessAction("call", "p1", map[string]interface{}{}, auth)
	if err != nil {
		t.Fatalf("Preflop call failed: %v", err)
	}

	// Still preflop? Wait, seat 1 called, current bet was 20.
	// In the real game, BB (seat 2) would have an option.
	// But our streetComplete logic checks if everyone ACTIVE has paid CurrentBet AND everyone has acted.
	// If seat 2 is BB, they haven't "acted" in the engine's view yet if we start with CurrentBet=20 but no ActedSeats.

	// Let's refine the state to be more realistic for the test.
	state.HandState.ActedSeats = []int{2} // BB already "acted" by posting
	state.Players[1].BetThisStreet = 20
	state.Players[1].ContributedThisHand = 20
	state.Players[0].BetThisStreet = 10 // SB
	state.Players[0].ContributedThisHand = 10
	state.HandState.Pot = 30
	state.HandState.ActingSeat = &[]int{1}[0] // SB's turn to call 10 more

	// 1. Preflop: SB calls
	_, err = be.ProcessAction("call", "p1", map[string]interface{}{}, auth)
	if err != nil {
		t.Fatalf("Preflop call failed: %v", err)
	}

	if state.HandState.Stage != "flop" {
		t.Errorf("Expected stage to be flop, got %s", state.HandState.Stage)
	}
	if len(state.HandState.CommunityCards) != 3 {
		t.Errorf("Expected 3 community cards, got %d", len(state.HandState.CommunityCards))
	}

	// 2. Flop: Check-Check
	// Action starts from first active seat after dealer (2).
	// Since dealer is 2, first active is 1.
	if *state.HandState.ActingSeat != 1 {
		t.Errorf("Expected acting seat 1, got %d", *state.HandState.ActingSeat)
	}

	_, err = be.ProcessAction("check", "p1", map[string]interface{}{}, auth)
	if err != nil {
		t.Fatalf("Flop check failed: %v", err)
	}
	_, err = be.ProcessAction("check", "p2", map[string]interface{}{}, auth)
	if err != nil {
		t.Fatalf("Flop second check failed: %v", err)
	}

	if state.HandState.Stage != "turn" {
		t.Errorf("Expected stage to be turn, got %s", state.HandState.Stage)
	}
	if len(state.HandState.CommunityCards) != 4 {
		t.Errorf("Expected 4 community cards, got %d", len(state.HandState.CommunityCards))
	}

	// 3. Turn: Bet-Fold
	_, err = be.ProcessAction("bet", "p1", map[string]interface{}{"amount": 100}, auth)
	if err != nil {
		t.Fatalf("Turn bet failed: %v", err)
	}
	showdown, err := be.ProcessAction("fold", "p2", map[string]interface{}{}, auth)
	if err != nil {
		t.Fatalf("Turn fold failed: %v", err)
	}

	if !showdown {
		t.Errorf("Expected fold to trigger showdown (end of hand) when only one player left")
	}
}

func TestBettingEngine_ComplexRaises(t *testing.T) {
	state := &models.TableState{
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", Stack: 1000, Name: "Alice", PlayerID: &[]string{"p1"}[0], BetThisStreet: 0},
			{Seat: 2, Status: "ACTIVE", Stack: 1000, Name: "Bob", PlayerID: &[]string{"p2"}[0], BetThisStreet: 0},
			{Seat: 3, Status: "ACTIVE", Stack: 1000, Name: "Charlie", PlayerID: &[]string{"p3"}[0], BetThisStreet: 0},
		},
		HandState: models.HandState{
			Status:       "in_progress",
			Stage:        "flop",
			CurrentBet:   0,
			ActingSeat:   &[]int{1}[0],
			MinimumRaise: BigBlind,
		},
	}
	log := func(msg string) { t.Log(msg) }
	be := NewBettingEngine(state, log)
	auth := func(seat int, pid string) bool { return true }

	// 1. Alice bets 40
	be.ProcessAction("bet", "p1", map[string]interface{}{"amount": 40}, auth)
	if state.HandState.CurrentBet != 40 {
		t.Errorf("Expected current bet 40, got %d", state.HandState.CurrentBet)
	}
	if state.HandState.MinimumRaise != 40 {
		t.Errorf("Expected min raise 40, got %d", state.HandState.MinimumRaise)
	}

	// 2. Bob raises to 120 (a raise of 80)
	be.ProcessAction("raise", "p2", map[string]interface{}{"amount": 120}, auth)
	if state.HandState.CurrentBet != 120 {
		t.Errorf("Expected current bet 120, got %d", state.HandState.CurrentBet)
	}
	if state.HandState.MinimumRaise != 80 {
		t.Errorf("Expected min raise 80, got %d", state.HandState.MinimumRaise)
	}

	// 3. Charlie raises to 300 (a raise of 180)
	be.ProcessAction("raise", "p3", map[string]interface{}{"amount": 300}, auth)
	if state.HandState.CurrentBet != 300 {
		t.Errorf("Expected current bet 300, got %d", state.HandState.CurrentBet)
	}
	if state.HandState.MinimumRaise != 180 {
		t.Errorf("Expected min raise 180, got %d", state.HandState.MinimumRaise)
	}

	// 4. Alice tries to raise to 400 (only 100 on top) - should fail because min raise is 180
	_, err := be.ProcessAction("raise", "p1", map[string]interface{}{"amount": 400}, auth)
	if err == nil {
		t.Errorf("Expected error for illegal raise amount (less than min raise)")
	}
}

func TestBettingEngine_AllInSidePots(t *testing.T) {
	state := &models.TableState{
		Players: []models.Player{
			{Seat: 1, Status: "ACTIVE", Stack: 100, Name: "Shorty", PlayerID: &[]string{"p1"}[0]},
			{Seat: 2, Status: "ACTIVE", Stack: 1000, Name: "Stacky1", PlayerID: &[]string{"p2"}[0]},
			{Seat: 3, Status: "ACTIVE", Stack: 1000, Name: "Stacky2", PlayerID: &[]string{"p3"}[0]},
		},
		HandState: models.HandState{
			Status:       "in_progress",
			Stage:        "flop",
			CurrentBet:   0,
			ActingSeat:   &[]int{1}[0],
			MinimumRaise: BigBlind,
			Pot:          0,
			Deck:         []string{"Ad", "Kd"},
		},
	}
	be := NewBettingEngine(state, func(string) {})
	auth := func(seat int, pid string) bool { return true }

	// Shorty goes all-in for 100
	be.ProcessAction("bet", "p1", map[string]interface{}{"amount": 100}, auth)
	if state.Players[0].Status != "ALL_IN" {
		t.Errorf("Expected player 1 to be ALL_IN")
	}

	// Stacky1 raises to 300
	be.ProcessAction("raise", "p2", map[string]interface{}{"amount": 300}, auth)
	
	// Stacky2 calls 300
	be.ProcessAction("call", "p3", map[string]interface{}{}, auth)

	if state.HandState.Pot != 700 { // 100 + 300 + 300
		t.Errorf("Expected pot 700, got %d", state.HandState.Pot)
	}

	// Check that we transitioned to turn (since Stacky1 and Stacky2 called/raised and everyone acted)
	if state.HandState.Stage != "turn" {
		t.Errorf("Expected stage to be turn, got %s", state.HandState.Stage)
	}

	// Check acting seat for turn - should be Stacky1 (seat 2) or Stacky2 (seat 3), but NOT Shorty (seat 1)
	if state.HandState.ActingSeat == nil || *state.HandState.ActingSeat == 1 {
		t.Errorf("Expected acting seat to be other than 1, got %v", state.HandState.ActingSeat)
	}
}
