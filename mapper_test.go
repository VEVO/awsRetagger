package main

import (
	"reflect"
	"regexp/syntax"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testData := []struct {
		input    string
		expected Mapper
	}{
		{"{}", Mapper{}},
		{
			`{
				"keys": [
						{"pattern": ".*production.*", "destination": [{"name": "Env", "value": "prd"}]},
  					{"pattern": ".*apache.*", "destination": [{"name": "Team", "value": "web"}, {"name": "Component", "value": "apache"}]}
					],
					"tags": [{"source": {"name": "Name", "value": ".*prod.*"}, "destination":[ {"name": "Env", "value": "prd"} ]}],
					"copy_tags": [{"sources": ["ENVIRONMENT", "ENVIRONMETNT", "Account"], "destination": "Env"}],
					"sanity": [{"tag_name": "Env", "remap": { "prd": ["prod", "production","global"], "stg": ["staging"], "dev": ["development"] }}],
					"defaults": {"Env": "unknown", "Team": "unknown", "Service": "unknown"}

			}`,
			Mapper{
				KeyMap: []*KeyMapper{
					{KeyPattern: ".*production.*", Destination: []*TagItem{{Name: "Env", Value: "prd"}}},
					{KeyPattern: ".*apache.*", Destination: []*TagItem{{Name: "Team", Value: "web"}, {Name: "Component", Value: "apache"}}},
				},
				TagMap:           []*TagMapper{{Source: &TagItem{Name: "Name", Value: ".*prod.*"}, Destination: []*TagItem{{Name: "Env", Value: "prd"}}}},
				CopyTag:          []*TagCopy{{Source: []string{"ENVIRONMENT", "ENVIRONMETNT", "Account"}, Destination: "Env"}},
				Sanity:           []*TagSanity{{TagName: "Env", Transform: map[string][]string{"prd": {"prod", "production", "global"}, "stg": {"staging"}, "dev": {"development"}}}},
				DefaultTagValues: map[string]string{"Env": "unknown", "Team": "unknown", "Service": "unknown"},
			},
		},
	}

	for _, d := range testData {
		m := Mapper{}
		if err := m.LoadConfig(strings.NewReader(d.input)); err != nil {
			t.Fatalf("LoadConfig returned: %s\n", err)
		}
		if !reflect.DeepEqual(d.expected, m) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, m)
		}
	}
}

func TestGetFromTags(t *testing.T) {
	testData := []struct {
		input, expected map[string]string
		config          Mapper
	}{
		{map[string]string{}, map[string]string{}, Mapper{}},
		{
			map[string]string{},
			map[string]string{},
			Mapper{
				CopyTag: []*TagCopy{
					{Source: []string{"ENVIRONMENT.*", "ENVIRONMETNT", "Account"}, Destination: "env"},
					{Source: []string{"unit", "band", "TeAm"}, Destination: "team"},
					{Source: []string{"application", "app", "project", "aplicaiton"}, Destination: "team"},
				},
			},
		},
		{
			map[string]string{"ENVIRONMENT-BLA": "regex bla", "Account": "should not be evaluated", "team": "already present", "unit": "should not be evaluated", "aplIcaitOn": "case insensitive match", "Name": "foo-bar-stg"},
			map[string]string{"env": "regex bla", "service": "case insensitive match", "component": "foo", "artist": "bar"},
			Mapper{
				CopyTag: []*TagCopy{
					{Source: []string{"ENVIRONMENT.*", "ENVIRONMETNT", "Account"}, Destination: "env"},
					{Source: []string{"unit", "band", "TeAm"}, Destination: "team"},
					{Source: []string{"application", "app", "project", "aplicaiton"}, Destination: "service"},
				},
				TagMap: []*TagMapper{
					{Source: &TagItem{Name: "Name", Value: ".*foo-.*stg.*"}, Destination: []*TagItem{{Name: "env", Value: "stg"}, {Name: "component", Value: "foo"}}},
					{Source: &TagItem{Name: "Name", Value: ".*foo.*"}, Destination: []*TagItem{{Name: "artist", Value: "bar"}, {Name: "component", Value: "foo that should not show"}}},
				},
			},
		},
	}

	for _, d := range testData {
		res, err := d.config.GetFromTags(&d.input)
		if err != nil {
			t.Fatalf("GetFromTags returned: %s\n", err)
		}
		if !reflect.DeepEqual(d.expected, *res) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, *res)
		}
	}
}

func TestGetFromKey(t *testing.T) {
	testData := []struct {
		inputKey            string
		inputTags, expected map[string]string
		config              Mapper
	}{
		{"", map[string]string{}, map[string]string{}, Mapper{}},
		{
			"nothing matches", map[string]string{}, map[string]string{},
			Mapper{
				KeyMap: []*KeyMapper{
					{KeyPattern: ".*-emr-.*", Destination: []*TagItem{{Name: "service", Value: "EMR"}}},
				},
			},
		},
		{
			"prod-emr-analytics", map[string]string{"random": "tag"},
			map[string]string{"service": "EMR", "environment": "prd", "team": "analytics"},
			Mapper{
				KeyMap: []*KeyMapper{
					{KeyPattern: ".*-emr-.*", Destination: []*TagItem{{Name: "service", Value: "EMR"}}},
					{KeyPattern: ".*prod.*", Destination: []*TagItem{{Name: "environment", Value: "prd"}}},
					{KeyPattern: ".*analytics.*", Destination: []*TagItem{{Name: "team", Value: "analytics"}, {Name: "service", Value: "should be ignored"}}},
					{KeyPattern: ".*foo.*", Destination: []*TagItem{{Name: "component", Value: "foo"}}},
				},
			},
		},
		{
			"prod-emr-analytics", map[string]string{"team": "data", "foo": "bar"},
			map[string]string{"service": "EMR", "environment": "prd"},
			Mapper{
				KeyMap: []*KeyMapper{
					{KeyPattern: ".*-emr-.*", Destination: []*TagItem{{Name: "service", Value: "EMR"}}},
					{KeyPattern: ".*prod.*", Destination: []*TagItem{{Name: "environment", Value: "prd"}}},
					{KeyPattern: ".*analytics.*", Destination: []*TagItem{{Name: "team", Value: "analytics"}, {Name: "service", Value: "should be ignored"}}},
					{KeyPattern: ".*foo.*", Destination: []*TagItem{{Name: "component", Value: "foo"}}},
				},
			},
		},
	}

	for _, d := range testData {
		res, err := d.config.GetFromKey(d.inputKey, &d.inputTags)
		if err != nil {
			t.Fatalf("GetFromKey returned: %s\n", err)
		}
		if !reflect.DeepEqual(d.expected, *res) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, *res)
		}
	}

}

func TestValidateTag(t *testing.T) {
	testData := []struct {
		tagName, tagValue string
		config            Mapper
		expectedTag       TagItem
		expectedError     error
	}{
		{"", "", Mapper{}, TagItem{Name: "", Value: ""}, NewErrSanityConfig("No sanity configuration found", "")},
		{"wrong_syntax", "in the regexp pattern", Mapper{
			Sanity: []*TagSanity{
				{TagName: "wrong_syntax", Transform: map[string][]string{"foo": {"ba)r"}}},
			},
		}, TagItem{Name: "wrong_syntax", Value: "in the regexp pattern"}, &syntax.Error{Code: syntax.ErrUnexpectedParen, Expr: "(?i)^ba)r$"}},
		{"foo", "bar", Mapper{
			Sanity: []*TagSanity{
				{TagName: "env", Transform: map[string][]string{"prd": {"prod.*", "global", "pdr"}, "stg": {"stag.*", "sgt"}, "dev": {"dev.*"}}},
				{TagName: "team", Transform: map[string][]string{"infrastructure": {"sys.*", "devops", "drunks"}, "web": {"frontend", "html"}, "ricard": {}}},
			},
		}, TagItem{Name: "foo", Value: "bar"}, NewErrSanityConfig("No sanity configuration found", "foo")},
		{"team", "drunks", Mapper{
			Sanity: []*TagSanity{
				{TagName: "env", Transform: map[string][]string{"prd": {"prod.*", "global", "pdr"}, "stg": {"stag.*", "sgt"}, "dev": {"dev.*"}}},
				{TagName: "team", Transform: map[string][]string{"infrastructure": {"sys.*", "devops", "drunks"}, "web": {"frontend", "html"}, "ricard": {}}},
			},
		}, TagItem{Name: "team", Value: "infrastructure"}, nil},
		{"env", "local", Mapper{
			Sanity: []*TagSanity{
				{TagName: "env", Transform: map[string][]string{"prd": {"prod.*", "global", "pdr"}, "stg": {"stag.*", "sgt"}, "dev": {"dev.*"}}},
				{TagName: "team", Transform: map[string][]string{"infrastructure": {"sys.*", "devops", "drunks"}, "web": {"frontend", "html"}, "ricard": {}}},
			},
		}, TagItem{Name: "env", Value: "local"}, NewErrSanityNoMapping("No match found for the sanity check", "env", "local")},
		{"env", "production", Mapper{
			Sanity: []*TagSanity{
				{TagName: "env", Transform: map[string][]string{"prd": {"prod.*", "global", "pdr"}, "stg": {"stag.*", "sgt"}, "dev": {"dev.*"}}},
				{TagName: "team", Transform: map[string][]string{"infrastructure": {"sys.*", "devops", "drunks"}, "web": {"frontend", "html"}, "ricard": {}}},
			},
		}, TagItem{Name: "env", Value: "prd"}, nil},
		{"team", "web", Mapper{
			Sanity: []*TagSanity{
				{TagName: "env", Transform: map[string][]string{"prd": {"prod.*", "global", "pdr"}, "stg": {"stag.*", "sgt"}, "dev": {"dev.*"}}},
				{TagName: "team", Transform: map[string][]string{"infrastructure": {"sys.*", "devops", "drunks"}, "web": {"frontend", "html"}, "ricard": {}}},
			},
		}, TagItem{Name: "team", Value: "web"}, nil},
		{"team", "ricard", Mapper{
			Sanity: []*TagSanity{
				{TagName: "env", Transform: map[string][]string{"prd": {"prod.*", "global", "pdr"}, "stg": {"stag.*", "sgt"}, "dev": {"dev.*"}}},
				{TagName: "team", Transform: map[string][]string{"infrastructure": {"sys.*", "devops", "drunks"}, "web": {"frontend", "html"}, "ricard": {}}},
			},
		}, TagItem{Name: "team", Value: "ricard"}, nil},
	}
	for _, d := range testData {
		res, err := d.config.ValidateTag(d.tagName, d.tagValue)
		if !reflect.DeepEqual(err, d.expectedError) {
			t.Errorf("GetFromKey returned: %s. Expecting: %s\n", err, d.expectedError)
		}
		if !reflect.DeepEqual(d.expectedTag, *res) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expectedTag, *res)
		}
	}
}

func TestStripDefaults(t *testing.T) {
	testData := []struct {
		input, expected map[string]string
		config          Mapper
	}{
		{map[string]string{}, map[string]string{}, Mapper{}},
		{map[string]string{"Env": "unknown", "owner": "Jack Skeleton"}, map[string]string{"Env": "unknown", "owner": "Jack Skeleton"}, Mapper{DefaultTagValues: map[string]string{}}},
		{map[string]string{"Env": "unknown", "owner": "Jack Skeleton"}, map[string]string{"owner": "Jack Skeleton"}, Mapper{DefaultTagValues: map[string]string{"Team": "unknown", "Env": "unknown"}}},
		{map[string]string{"Env": "prod", "owner": "Jack Skeleton"}, map[string]string{"Env": "prod", "owner": "Jack Skeleton"}, Mapper{DefaultTagValues: map[string]string{"Team": "unknown", "Env": "unknown"}}},
	}
	for _, d := range testData {
		d.config.StripDefaults(&d.input)
		if !reflect.DeepEqual(d.expected, d.input) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, d.input)
		}
	}
}

func TestGetMissingDefaults(t *testing.T) {
	testData := []struct {
		input, expected map[string]string
		config          Mapper
	}{
		{map[string]string{}, map[string]string{}, Mapper{}},
		{map[string]string{"Env": "prod", "owner": "Jack Skeleton"}, map[string]string{}, Mapper{DefaultTagValues: map[string]string{}}},
		{map[string]string{"Env": "prod", "owner": "Jack Skeleton"}, map[string]string{"Team": "unknown"}, Mapper{DefaultTagValues: map[string]string{"Team": "unknown", "Env": "unknown"}}},
		{map[string]string{"Env": "unknown", "owner": "Jack Skeleton"}, map[string]string{"Team": "unknown"}, Mapper{DefaultTagValues: map[string]string{"Team": "unknown", "Env": "unknown"}}},
	}
	for _, d := range testData {
		res := d.config.GetMissingDefaults(&d.input)
		if !reflect.DeepEqual(d.expected, *res) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, *res)
		}
	}
}
