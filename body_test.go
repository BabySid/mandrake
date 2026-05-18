package main

import (
	"reflect"
	"testing"
)

func TestBodySetTopLevel(t *testing.T) {
	body := map[string]any{"name": "alice"}
	BodySet(body, "age", float64(30))
	if body["age"] != float64(30) {
		t.Errorf("age = %v, want 30", body["age"])
	}
}

func TestBodySetNested(t *testing.T) {
	body := map[string]any{}
	BodySet(body, "data.user.name", "bob")
	data := body["data"].(map[string]any)
	user := data["user"].(map[string]any)
	if user["name"] != "bob" {
		t.Errorf("data.user.name = %v, want bob", user["name"])
	}
}

func TestBodySetOverwriteExisting(t *testing.T) {
	body := map[string]any{
		"data": map[string]any{
			"model": "old",
		},
	}
	BodySet(body, "data.model", "new")
	data := body["data"].(map[string]any)
	if data["model"] != "new" {
		t.Errorf("data.model = %v, want new", data["model"])
	}
}

func TestBodySetAutoCreateIntermediate(t *testing.T) {
	body := map[string]any{}
	BodySet(body, "a.b.c.d", "deep")
	a := body["a"].(map[string]any)
	b := a["b"].(map[string]any)
	c := b["c"].(map[string]any)
	if c["d"] != "deep" {
		t.Errorf("a.b.c.d = %v, want deep", c["d"])
	}
}

func TestBodyDeleteTopLevel(t *testing.T) {
	body := map[string]any{"name": "alice", "age": float64(30)}
	BodyDelete(body, "age")
	if _, ok := body["age"]; ok {
		t.Error("age should be deleted")
	}
	if body["name"] != "alice" {
		t.Error("name should remain")
	}
}

func TestBodyDeleteNested(t *testing.T) {
	body := map[string]any{
		"data": map[string]any{
			"user": map[string]any{
				"name": "alice",
				"age":  float64(30),
			},
		},
	}
	BodyDelete(body, "data.user.age")
	user := body["data"].(map[string]any)["user"].(map[string]any)
	if _, ok := user["age"]; ok {
		t.Error("data.user.age should be deleted")
	}
	if user["name"] != "alice" {
		t.Error("data.user.name should remain")
	}
}

func TestBodyDeleteMissingPath(t *testing.T) {
	body := map[string]any{"name": "alice"}
	BodyDelete(body, "nonexistent.deep.path")
	if body["name"] != "alice" {
		t.Error("name should remain")
	}
}

func TestBodyDeleteIntermediateNotObject(t *testing.T) {
	body := map[string]any{"name": "alice"}
	BodyDelete(body, "name.sub.field")
	if body["name"] != "alice" {
		t.Error("name should remain unchanged")
	}
}

func TestModifyBody(t *testing.T) {
	body := map[string]any{
		"model":    "old-model",
		"internal": "secret",
	}
	mods := BodyMods{
		Set:    map[string]any{"model": "new-model", "meta.source": "mandrake"},
		Delete: []string{"internal"},
	}

	ModifyBody(body, mods)

	if body["model"] != "new-model" {
		t.Errorf("model = %v, want new-model", body["model"])
	}
	if _, ok := body["internal"]; ok {
		t.Error("internal should be deleted")
	}
	meta := body["meta"].(map[string]any)
	if meta["source"] != "mandrake" {
		t.Errorf("meta.source = %v, want mandrake", meta["source"])
	}
}

func TestModifyBodyEmpty(t *testing.T) {
	body := map[string]any{"key": "value"}
	orig := map[string]any{"key": "value"}
	ModifyBody(body, BodyMods{})
	if !reflect.DeepEqual(body, orig) {
		t.Error("empty mods should not change body")
	}
}

func TestBodyPrepend(t *testing.T) {
	body := map[string]any{
		"system": []any{
			map[string]any{"type": "text", "text": "existing"},
		},
	}
	BodyPrepend(body, "system", map[string]any{"type": "text", "text": "first"})
	arr := body["system"].([]any)
	if len(arr) != 2 {
		t.Fatalf("system len = %d, want 2", len(arr))
	}
	first := arr[0].(map[string]any)
	if first["text"] != "first" {
		t.Errorf("system[0].text = %v, want first", first["text"])
	}
	second := arr[1].(map[string]any)
	if second["text"] != "existing" {
		t.Errorf("system[1].text = %v, want existing", second["text"])
	}
}

func TestBodyPrependMissingArray(t *testing.T) {
	body := map[string]any{"name": "alice"}
	BodyPrepend(body, "system", map[string]any{"type": "text"})
	if _, ok := body["system"]; ok {
		t.Error("should not create array if it doesn't exist")
	}
}

func TestBodyAppend(t *testing.T) {
	body := map[string]any{
		"system": []any{
			map[string]any{"type": "text", "text": "existing"},
		},
	}
	BodyAppend(body, "system", map[string]any{"type": "text", "text": "last"})
	arr := body["system"].([]any)
	if len(arr) != 2 {
		t.Fatalf("system len = %d, want 2", len(arr))
	}
	last := arr[1].(map[string]any)
	if last["text"] != "last" {
		t.Errorf("system[1].text = %v, want last", last["text"])
	}
}
