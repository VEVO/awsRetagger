package mapper

import (
	"io"
)

// PutTagFn is used to specify the function structure to pass to the Retag method
type PutTagFn func(*string, []*TagItem) error

// Iface has been created for testing purposes. It allows to create mocks
// when testing class that depend on the mapper
type Iface interface {
	LoadConfig(io.Reader) error
	StripDefaults(*map[string]string)
	GetMissingDefaults(*map[string]string) *map[string]string
	GetFromTags(*map[string]string) (*map[string]string, error)
	GetFromKey(string, *map[string]string) (*map[string]string, error)
	ValidateTag(string, string) (*TagItem, error)
	MergeMaps(*map[string]string, *map[string]string)

	Retag(*string, *map[string]string, []string, PutTagFn)
}

var _ Iface = (*Mapper)(nil)
