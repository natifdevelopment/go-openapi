package merger

// Conflict represents a merge conflict between services.
type Conflict struct {
	Type     string `json:"type"`
	Service  string `json:"service"`
	Path     string `json:"path,omitempty"`
	Detail   string `json:"detail"`
	Severity string `json:"severity"`
}

// addConflict records a merge conflict.
func (m *Merger) addConflict(conflictType, serviceKey, path, detail string) {
	severity := "error"
	// Check if this conflict type is in the warn_on list
	for _, w := range m.Config.Validation.WarnOn {
		if w == conflictType {
			severity = "warning"
			break
		}
	}

	c := Conflict{
		Type:     conflictType,
		Service:  serviceKey,
		Path:     path,
		Detail:   detail,
		Severity: severity,
	}

	if severity == "warning" {
		m.warnings = append(m.warnings, ValidationResult{
			Rule:     conflictType,
			Severity: "warning",
			Path:     path,
			Message:  detail,
		})
	} else {
		m.conflicts = append(m.conflicts, c)
	}
}
