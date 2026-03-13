package game

import (
	"fmt"
	"sync"
	"testing"
)

func TestTable_Concurrency(t *testing.T) {
	tbl := NewTable("concurrency-table", false)

	// Add 6 players
	for i := 1; i <= 6; i++ {
		pid := fmt.Sprintf("p%d", i)
		tbl.Join(pid, fmt.Sprintf("Player %d", i))
		tbl.ApplyAction("join_game", map[string]interface{}{"player_id": pid, "seat": i, "player_name": fmt.Sprintf("Player %d", i)})
	}

	// Start a hand
	tbl.ApplyAction("next_hand", map[string]interface{}{"player_id": "1"})

	var wg sync.WaitGroup
	numActions := 100
	wg.Add(numActions)

	// Simulate concurrent actions (some might be invalid depending on turn, but the lock should prevent crashes)
	for i := range numActions {
		go func(idx int) {
			defer wg.Done()
			pid := fmt.Sprintf("p%d", (idx%6)+1)
			action := "call"
			if idx%3 == 0 {
				action = "check"
			} else if idx%5 == 0 {
				action = "raise"
			}

			tbl.ApplyAction(action, map[string]interface{}{
				"player_id": pid,
				"amount":    100,
			})
		}(i)
	}

	wg.Wait()

	// If we reached here without a crash or deadlock, the basic locking is working.
	state := tbl.GetState()
	if state.TableID != "concurrency-table" {
		t.Errorf("State corrupted?")
	}
}
