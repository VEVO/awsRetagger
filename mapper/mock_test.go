package mapper

import (
	"reflect"
	"testing"
)

func dummyMockMapperFct(res *string, tag []*TagItem) error {
	return nil
}

func TestMockRetag(t *testing.T) {
	testData := []struct {
		inputResourceTags, outputResourceTags map[string]map[string]string
		inputResourceKeys, outputResourceKeys map[string][]string
	}{
		{map[string]map[string]string{}, nil, map[string][]string{}, nil},
		{
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string][]string{"bar": {"mykey1", "mykey2"}, "joe": {"mykey3"}},
			map[string][]string{"foo": nil, "bar": {"mykey1", "mykey2"}, "joe": {"mykey3"}},
		},
	}

	for _, d := range testData {
		m := MockMapper{}
		for k, v := range d.inputResourceTags {
			inKeys := []string{}
			inKeys, _ = d.inputResourceKeys[k]
			m.Retag(&k, &v, inKeys, nil)
		}
		if !reflect.DeepEqual(d.outputResourceTags, m.ResourceTags) {
			t.Errorf("Expecting ResourceTags: %v\nGot: %v\n", d.outputResourceTags, m.ResourceTags)
		}

		if !reflect.DeepEqual(d.outputResourceKeys, m.ResourceKeys) {
			t.Errorf("Expecting ResourceKeys: %v\nGot: %v\n", d.outputResourceKeys, m.ResourceKeys)
		}
	}
}
