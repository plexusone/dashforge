package interaction

import (
	"context"
	"testing"

	"github.com/plexusone/uiforge/pkg/state"
	"github.com/plexusone/uiforge/uispec"
)

func TestDispatchStateSet(t *testing.T) {
	store := state.New()
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "chart", Event: "click"},
			Then: []uispec.InteractionAction{
				{
					Action: "state.set",
					Params: map[string]any{"path": "selected.month", "value": "July"},
				},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "chart",
		EventName:   "click",
		Data:        map[string]any{"month": "July"},
	})
	if err != nil {
		t.Fatal(err)
	}

	val, ok := store.Get("selected.month")
	if !ok {
		t.Fatal("expected state to be set")
	}
	if val != "July" {
		t.Errorf("got %v, want July", val)
	}
}

func TestDispatchExpressionResolution(t *testing.T) {
	store := state.New()
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "chart", Event: "pointSelected"},
			Then: []uispec.InteractionAction{
				{
					Action: "state.set",
					Params: map[string]any{
						"path":  "filters.month",
						"value": "${event.month}",
					},
				},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "chart",
		EventName:   "pointSelected",
		Data:        map[string]any{"month": "March"},
	})
	if err != nil {
		t.Fatal(err)
	}

	val, ok := store.Get("filters.month")
	if !ok {
		t.Fatal("expected state to be set")
	}
	if val != "March" {
		t.Errorf("got %v, want March", val)
	}
}

func TestDispatchNoMatch(t *testing.T) {
	store := state.New()
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "chart", Event: "click"},
			Then: []uispec.InteractionAction{
				{Action: "state.set", Params: map[string]any{"path": "x", "value": 1}},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "table",
		EventName:   "click",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, ok := store.Get("x")
	if ok {
		t.Error("no interaction should have matched")
	}
}

func TestDispatchConditionTrue(t *testing.T) {
	store := state.New()
	store.Set("admin", true)
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "btn", Event: "click"},
			Then: []uispec.InteractionAction{
				{
					Action:    "state.set",
					Condition: "${admin}",
					Params:    map[string]any{"path": "result", "value": "ok"},
				},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "btn",
		EventName:   "click",
	})
	if err != nil {
		t.Fatal(err)
	}

	val, ok := store.Get("result")
	if !ok {
		t.Fatal("expected state to be set when condition is true")
	}
	if val != "ok" {
		t.Errorf("got %v, want ok", val)
	}
}

func TestDispatchConditionFalse(t *testing.T) {
	store := state.New()
	store.Set("admin", false)
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "btn", Event: "click"},
			Then: []uispec.InteractionAction{
				{
					Action:    "state.set",
					Condition: "${admin}",
					Params:    map[string]any{"path": "result", "value": "ok"},
				},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "btn",
		EventName:   "click",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, ok := store.Get("result")
	if ok {
		t.Error("expected state NOT to be set when condition is false")
	}
}

func TestDispatchToggle(t *testing.T) {
	store := state.New()
	store.Set("sidebar.open", true)
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "toggle", Event: "click"},
			Then: []uispec.InteractionAction{
				{Action: "state.toggle", Params: map[string]any{"path": "sidebar.open"}},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "toggle",
		EventName:   "click",
	})
	if err != nil {
		t.Fatal(err)
	}

	val, _ := store.Get("sidebar.open")
	if val != false {
		t.Errorf("got %v, want false", val)
	}
}

func TestDispatchUnknownAction(t *testing.T) {
	store := state.New()
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "btn", Event: "click"},
			Then: []uispec.InteractionAction{
				{Action: "nonexistent.action"},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "btn",
		EventName:   "click",
	})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestDispatchCustomHandler(t *testing.T) {
	store := state.New()
	eng := New(store)

	var called bool
	eng.RegisterHandler("custom.action", func(_ context.Context, action uispec.InteractionAction, _ map[string]any) error {
		called = true
		if action.Target != "my-target" {
			t.Errorf("target = %q, want my-target", action.Target)
		}
		return nil
	})

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "btn", Event: "click"},
			Then: []uispec.InteractionAction{
				{Action: "custom.action", Target: "my-target"},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "btn",
		EventName:   "click",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("custom handler was not called")
	}
}

func TestDispatchMultipleActions(t *testing.T) {
	store := state.New()
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "chart", Event: "select"},
			Then: []uispec.InteractionAction{
				{Action: "state.set", Params: map[string]any{"path": "a", "value": 1}},
				{Action: "state.set", Params: map[string]any{"path": "b", "value": 2}},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "chart",
		EventName:   "select",
	})
	if err != nil {
		t.Fatal(err)
	}

	a, _ := store.Get("a")
	b, _ := store.Get("b")
	if a != 1 || b != 2 {
		t.Errorf("got a=%v b=%v, want 1 2", a, b)
	}
}

func TestDispatchComponentRefresh(t *testing.T) {
	store := state.New()
	eng := New(store)

	interactions := []uispec.Interaction{
		{
			When: uispec.InteractionTrigger{Component: "filter", Event: "change"},
			Then: []uispec.InteractionAction{
				{Action: "component.refresh", Target: "table"},
			},
		},
	}

	err := eng.Dispatch(context.Background(), interactions, Event{
		ComponentID: "filter",
		EventName:   "change",
	})
	if err != nil {
		t.Fatal(err)
	}
}
