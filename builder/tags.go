package builder

import "sort"

// TagRegistry manages OpenAPI tags during spec building.
type TagRegistry struct {
	tags map[string]map[string]interface{} // name → tag map
}

// NewTagRegistry creates a new TagRegistry.
func NewTagRegistry() *TagRegistry {
	return &TagRegistry{tags: make(map[string]map[string]interface{})}
}

// Register adds a tag if it doesn't already exist.
func (r *TagRegistry) Register(name, description string) {
	if existing, ok := r.tags[name]; ok {
		if len(description) > len(existing["description"].(string)) {
			existing["description"] = description
		}
		return
	}
	r.tags[name] = map[string]interface{}{
		"name":        name,
		"description": description,
	}
}

// Sorted returns the tags sorted alphabetically by name as []interface{}.
func (r *TagRegistry) Sorted() []interface{} {
	names := make([]string, 0, len(r.tags))
	for name := range r.tags {
		names = append(names, name)
	}
	sort.Strings(names)

	tags := make([]interface{}, 0, len(names))
	for _, name := range names {
		tags = append(tags, r.tags[name])
	}
	return tags
}
