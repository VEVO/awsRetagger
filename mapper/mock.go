package mapper

// MockMapper is used to mock the calls to retag during the tests
type MockMapper struct {
	Iface
	// ResourceTags is used to record which tags have been pushed to the Retag
	// function on which resource since the creation of the object
	ResourceTags map[string]map[string]string
	// ResourceKeys is used to record which keys have been pushed to the
	// Retag function on which resource since the creation of the object
	ResourceKeys map[string][]string
}

// Retag just records which resource has been called with which tags
func (m *MockMapper) Retag(resourceID *string, tags *map[string]string, keys []string, setTags PutTagFn) {
	if m.ResourceTags == nil {
		m.ResourceTags = make(map[string]map[string]string)
	}
	if m.ResourceKeys == nil {
		m.ResourceKeys = make(map[string][]string)
	}
	if val, ok := m.ResourceTags[*resourceID]; ok {
		for k, v := range *tags {
			if _, exists := val[k]; !exists {
				m.ResourceTags[*resourceID][k] = v
			}
		}
	} else {
		m.ResourceTags[*resourceID] = *tags
	}
	if val, ok := m.ResourceKeys[*resourceID]; ok {
		m.ResourceKeys[*resourceID] = append(val, keys...)
	} else {
		m.ResourceKeys[*resourceID] = keys
	}
}
