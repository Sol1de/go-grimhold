package world

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

var worldJSON []byte

type Region struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Danger      int      `json:"danger"`
	Foes        []string `json:"foes"`
	Exits       []string `json:"exits"`
}

type World struct {
	Start   string   `json:"start"`
	Regions []Region `json:"regions"`
}

func Load() (*World, error) { return parse(worldJSON) }

func parse(data []byte) (*World, error) {
	var w World
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("world.json invalide : %w", err)
	}
	seen := map[string]bool{}
	for _, r := range w.Regions {
		if r.ID == "" {
			return nil, errors.New("world.json : région sans id")
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("world.json : région %q en double", r.ID)
		}
		seen[r.ID] = true
	}
	if !seen[w.Start] {
		return nil, fmt.Errorf("world.json : région de départ %q inconnue", w.Start)
	}
	for _, r := range w.Regions {
		for _, e := range r.Exits {
			if !seen[e] {
				return nil, fmt.Errorf("world.json : sortie %q de %q inconnue", e, r.ID)
			}
		}
	}
	return &w, nil
}

func (w *World) Region(id string) (Region, bool) {
	i := slices.IndexFunc(w.Regions, func(r Region) bool { return r.ID == id })
	if i < 0 {
		return Region{}, false
	}
	return w.Regions[i], true
}

func (r Region) HasExit(id string) bool { return slices.Contains(r.Exits, id) }
