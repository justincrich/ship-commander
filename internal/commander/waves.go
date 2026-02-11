package commander

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ComputeWaves topologically sorts missions into dependency-safe wave batches.
func ComputeWaves(missions []Mission) ([][]Mission, error) {
	if len(missions) == 0 {
		return [][]Mission{}, nil
	}

	byID := make(map[string]Mission, len(missions))
	index := make(map[string]int, len(missions))
	for i, mission := range missions {
		if strings.TrimSpace(mission.ID) == "" {
			return nil, fmt.Errorf("mission at index %d has empty id", i)
		}
		if _, exists := byID[mission.ID]; exists {
			return nil, fmt.Errorf("duplicate mission id %q", mission.ID)
		}
		byID[mission.ID] = mission
		index[mission.ID] = i
	}

	indegree := make(map[string]int, len(missions))
	children := make(map[string][]string, len(missions))
	for _, mission := range missions {
		indegree[mission.ID] = 0
	}

	for _, mission := range missions {
		for _, dep := range mission.DependsOn {
			if _, ok := byID[dep]; !ok {
				continue
			}
			indegree[mission.ID]++
			children[dep] = append(children[dep], mission.ID)
		}
	}

	current := make([]string, 0, len(missions))
	for _, mission := range missions {
		if indegree[mission.ID] == 0 {
			current = append(current, mission.ID)
		}
	}

	visited := 0
	waves := make([][]Mission, 0)
	for len(current) > 0 {
		sort.SliceStable(current, func(i, j int) bool {
			return index[current[i]] < index[current[j]]
		})

		wave := make([]Mission, 0, len(current))
		next := make([]string, 0)
		for _, id := range current {
			wave = append(wave, byID[id])
			visited++
			for _, child := range children[id] {
				indegree[child]--
				if indegree[child] == 0 {
					next = append(next, child)
				}
			}
		}

		waves = append(waves, wave)
		current = next
	}

	if visited != len(missions) {
		return nil, fmt.Errorf("dependency cycle detected among missions")
	}

	return waves, nil
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = nonSlugChars.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "mission"
	}
	return normalized
}

func missionToken(missionID string) string {
	token := slugify(missionID)
	if token == "mission" {
		return "MISSION-unknown"
	}
	return "MISSION-" + token
}
