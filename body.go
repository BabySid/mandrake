package main

import "strings"

func BodySet(obj map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := obj
	for _, key := range parts[:len(parts)-1] {
		next, ok := current[key]
		if !ok {
			m := map[string]any{}
			current[key] = m
			current = m
			continue
		}
		m, ok := next.(map[string]any)
		if !ok {
			m = map[string]any{}
			current[key] = m
			current = m
			continue
		}
		current = m
	}
	current[parts[len(parts)-1]] = value
}

func BodyDelete(obj map[string]any, path string) {
	parts := strings.Split(path, ".")
	current := obj
	for _, key := range parts[:len(parts)-1] {
		next, ok := current[key]
		if !ok {
			return
		}
		m, ok := next.(map[string]any)
		if !ok {
			return
		}
		current = m
	}
	delete(current, parts[len(parts)-1])
}

func BodyPrepend(obj map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := obj
	for _, key := range parts[:len(parts)-1] {
		next, ok := current[key]
		if !ok {
			return
		}
		m, ok := next.(map[string]any)
		if !ok {
			return
		}
		current = m
	}
	key := parts[len(parts)-1]
	arr, ok := current[key].([]any)
	if !ok {
		return
	}
	current[key] = append([]any{value}, arr...)
}

func BodyAppend(obj map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := obj
	for _, key := range parts[:len(parts)-1] {
		next, ok := current[key]
		if !ok {
			return
		}
		m, ok := next.(map[string]any)
		if !ok {
			return
		}
		current = m
	}
	key := parts[len(parts)-1]
	arr, ok := current[key].([]any)
	if !ok {
		return
	}
	current[key] = append(arr, value)
}

func ModifyBody(obj map[string]any, mods BodyMods) {
	for path, value := range mods.Set {
		BodySet(obj, path, value)
	}
	for _, path := range mods.Delete {
		BodyDelete(obj, path)
	}
	for path, value := range mods.Prepend {
		BodyPrepend(obj, path, value)
	}
	for path, value := range mods.Append {
		BodyAppend(obj, path, value)
	}
}
