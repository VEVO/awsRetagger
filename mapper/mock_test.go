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
		existingTags, inputResourceTags, outputResourceTags map[string]map[string]string
		existingKeys, inputResourceKeys, outputResourceKeys map[string][]string
	}{
		{nil, map[string]map[string]string{}, nil, nil, map[string][]string{}, nil},
		{
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string]map[string]string{},
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string][]string{"joe": {"mykey4"}, "bar": {"key3"}},
			map[string][]string{},
			map[string][]string{"joe": {"mykey4"}, "bar": {"key3"}},
		},
		{
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string]map[string]string{"foo": {"tofoo1": "spicy1"}},
			map[string]map[string]string{"foo": {"key1": "value1", "tofoo1": "spicy1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string][]string{"joe": {"mykey4"}, "bar": {"key6"}},
			map[string][]string{"joe": {"mykey4"}, "bar": {"key3", "key8"}},
			map[string][]string{"foo": nil, "joe": {"mykey4", "mykey4"}, "bar": {"key6", "key3", "key8"}},
		},
		{
			nil,
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			map[string]map[string]string{"foo": {"key1": "value1"}, "bar": {"key2": "value2", "key3": "value3"}, "joe": nil},
			nil,
			map[string][]string{"bar": {"mykey1", "mykey2"}, "joe": {"mykey3"}},
			map[string][]string{"foo": nil, "bar": {"mykey1", "mykey2"}, "joe": {"mykey3"}},
		},
	}

	for _, d := range testData {
		m := MockMapper{ResourceTags: d.existingTags, ResourceKeys: d.existingKeys}
		resources := []string{}
		for k := range d.inputResourceTags {
			resources = append(resources, k)
		}
		for k := range d.inputResourceKeys {
			// Only add if not already there
			found := false
			for _, rez := range resources {
				if rez == k {
					found = true
					break
				}
			}
			if !found {
				resources = append(resources, k)
			}
		}
		for _, k := range resources {
			v, _ := d.inputResourceTags[k]
			inKeys, _ := d.inputResourceKeys[k]
			t.Logf("%s, %v, %v", k, v, inKeys)
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
