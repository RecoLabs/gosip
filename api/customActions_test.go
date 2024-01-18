package api

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCustomActions(t *testing.T) {
	checkClient(t)

	sp := NewSP(spClient)

	t.Run("Get/Site", func(t *testing.T) {
		actions, err := sp.Site().CustomActions().Top(1).Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(actions) > 0 {
			if actions[0].ID == "" {
				t.Error("can't get custom action data")
			}
		}
	})

	t.Run("Get/Web", func(t *testing.T) {
		actions, err := sp.Web().CustomActions().Top(1).Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(actions) > 0 {
			if actions[0].ID == "" {
				t.Error("can't get custom action data")
			}
		}
	})

	t.Run("Action/AddGetDelete", func(t *testing.T) {
		info := map[string]interface{}{
			"Location":    "ScriptLink",
			"Sequence":    100,
			"ScriptBlock": "if (console) { console.log(1); }",
		}
		// Add
		payload, _ := json.Marshal(info)
		action, err := sp.Web().CustomActions().Add(context.Background(), payload)
		if err != nil {
			t.Error(err)
		}
		// Get
		action1, err := sp.Web().CustomActions().GetByID(action.ID).Get(context.Background())
		if err != nil {
			t.Error(err)
		}
		if action.ID != action1.ID {
			t.Error("can't get action by ID")
		}
		// Delete
		if err := sp.Web().CustomActions().GetByID(action.ID).Delete(context.Background()); err != nil {
			t.Error(err)
		}
	})

}
