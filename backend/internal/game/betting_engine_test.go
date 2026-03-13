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
