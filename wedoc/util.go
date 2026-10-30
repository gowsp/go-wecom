package wedoc

import (
	"errors"
	"strings"
)

func normalizeUserIDs(userIDs []string) []string {
	result := make([]string, 0, len(userIDs))
	seen := make(map[string]bool, len(userIDs))
	for _, userID := range userIDs {
		id := strings.TrimSpace(userID)
		if id != "" && !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// errEmpty reports a required list that carries no entries.
func errEmpty(op, field string) error {
	return errors.New("wecom: " + op + ": " + field + " is required")
}
