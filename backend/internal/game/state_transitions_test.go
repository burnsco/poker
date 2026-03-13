package game

import (
	"testing"
)

func TestTable_MaterializePendingPlayers(t *testing.T) {
	tbl := NewTable("test-table", false)
	
	// Alice is already at the table
	tbl.Join("p1", "Alice")
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": "p1", "seat": 1, "player_name": "Alice"})
	
	// Bob joins while a hand is "in progress"
	tbl.state.HandState.Status = "in_progress"
	tbl.Join("p2", "Bob")
	tbl.ApplyAction("join_game", map[string]interface{}{"player_id": "p2", "seat": 2, "player_name": "Bob"})
	
	// Because seat 2 was empty, Bob claims it immediately but status is SITTING_OUT
	if len(tbl.state.PendingPlayers) != 0 {
		t.Errorf("Expected 0 pending players (claiming empty seat), got %d", len(tbl.state.PendingPlayers))
	}
	if tbl.state.Players[1].Name != "Bob" {
		t.Errorf("Expected Bob at seat 2, got %s", tbl.state.Players[1].Name)
	}
	if tbl.state.Players[1].Status != "SITTING_OUT" {
		t.Errorf("Expected Bob to be SITTING_OUT while hand in progress, got %s", tbl.state.Players[1].Status)
	}

	// End hand and start next
	tbl.state.HandState.Status = "complete"
	tbl.ApplyAction("next_hand", map[string]interface{}{"player_id": "1"})

	if tbl.state.Players[1].Status != "ACTIVE" {
		t.Errorf("Expected Bob to be ACTIVE for next hand, got %s", tbl.state.Players[1].Status)
	}
}
