package main

import (
	"errors"
	"github.com/sirupsen/logrus"
	logrus_test "github.com/sirupsen/logrus/hooks/test"
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
		expectedError   error
	}{
		{map[string]string{}, map[string]string{}, Mapper{}, nil},
		{
			map[string]string{},
			map[string]string{},
			Mapper{
				CopyTag: []*TagCopy{
					{Source: []string{"ENVIRONMENT.*", "ENVIRONMETNT", "Account"}, Destination: "env"},
					{Source: []string{"unit", "band", "TeAm"}, Destination: "team"},
					{Source: []string{"application", "app", "project", "aplicaiton"}, Destination: "team"},
				},
			}, nil,
		},
		{
			map[string]string{"ENVIRONMENT-BLA": "regex bla", "Account": "should not be evaluated", "team": "already present", "unit": "should not be evaluated", "aplIcaitOn": "case insensitive match", "Name": "foo-bar-stg"},
			map[string]string{},
			Mapper{
				CopyTag: []*TagCopy{
					{Source: []string{"EN)VIRONMENT.*", "ENVIRONMETNT", "Account"}, Destination: "env"},
					{Source: []string{"unit", "band", "TeAm"}, Destination: "team"},
					{Source: []string{"application", "app", "project", "aplicaiton"}, Destination: "team"},
				},
			}, &syntax.Error{Code: syntax.ErrUnexpectedParen, Expr: "(?i)^EN)VIRONMENT.*$"},
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
			}, nil,
		},
		{
			map[string]string{"ENVIRONMENT-BLA": "regex bla", "Account": "should not be evaluated", "team": "already present", "unit": "should not be evaluated", "aplIcaitOn": "case insensitive match", "Name": "foo-bar-stg"},
			map[string]string{"env": "regex bla", "service": "case insensitive match"},
			Mapper{
				CopyTag: []*TagCopy{
					{Source: []string{"ENVIRONMENT.*", "ENVIRONMETNT", "Account"}, Destination: "env"},
					{Source: []string{"unit", "band", "TeAm"}, Destination: "team"},
					{Source: []string{"application", "app", "project", "aplicaiton"}, Destination: "service"},
				},
				TagMap: []*TagMapper{
					{Source: &TagItem{Name: "Name", Value: ".*foo-.*s)tg.*"}, Destination: []*TagItem{{Name: "env", Value: "stg"}, {Name: "component", Value: "foo"}}},
					{Source: &TagItem{Name: "Name", Value: ".*foo.*"}, Destination: []*TagItem{{Name: "artist", Value: "bar"}, {Name: "component", Value: "foo that should not show"}}},
				},
			}, &syntax.Error{Code: syntax.ErrUnexpectedParen, Expr: "(?i)^.*foo-.*s)tg.*$"},
		},
	}

	for _, d := range testData {
		res, err := d.config.GetFromTags(&d.input)
		if !reflect.DeepEqual(err, d.expectedError) {
			t.Fatalf("GetFromTags returned: %v, expecting: %v\n", err, d.expectedError)
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
		expectedError       error
	}{
		{"", map[string]string{}, map[string]string{}, Mapper{}, nil},
		{
			"nothing matches", map[string]string{}, map[string]string{},
			Mapper{
				KeyMap: []*KeyMapper{
					{KeyPattern: ".*-emr-.*", Destination: []*TagItem{{Name: "service", Value: "EMR"}}},
				},
			}, nil,
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
			}, nil,
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
			}, nil,
		},
		{
			"prod-emr-analytics", map[string]string{"team": "data", "foo": "bar"},
			map[string]string{"service": "EMR"},
			Mapper{
				KeyMap: []*KeyMapper{
					{KeyPattern: ".*-emr-.*", Destination: []*TagItem{{Name: "service", Value: "EMR"}}},
					{KeyPattern: ".*p)rod.*", Destination: []*TagItem{{Name: "environment", Value: "prd"}}},
					{KeyPattern: ".*analytics.*", Destination: []*TagItem{{Name: "team", Value: "analytics"}, {Name: "service", Value: "should be ignored"}}},
					{KeyPattern: ".*foo.*", Destination: []*TagItem{{Name: "component", Value: "foo"}}},
				},
			}, &syntax.Error{Code: syntax.ErrUnexpectedParen, Expr: "(?i)^.*p)rod.*$"},
		},
	}

	for _, d := range testData {
		res, err := d.config.GetFromKey(d.inputKey, &d.inputTags)
		if !reflect.DeepEqual(err, d.expectedError) {
			t.Fatalf("GetFromKey returned: %v, expecting: %v\n", err, d.expectedError)
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

func TestMergeMaps(t *testing.T) {
	testData := []struct {
		mainMap, additionalMap, expected map[string]string
	}{
		{map[string]string{}, map[string]string{}, map[string]string{}},
		{map[string]string{"foo": "bar"}, map[string]string{}, map[string]string{"foo": "bar"}},
		{map[string]string{}, map[string]string{"foo": "bar"}, map[string]string{"foo": "bar"}},
		{map[string]string{"foo": "bar"}, map[string]string{"foo": "already there!"}, map[string]string{"foo": "bar"}},
		{map[string]string{"foo": "bar"}, map[string]string{"foo": "already there!", "hello": "world!"}, map[string]string{"foo": "bar", "hello": "world!"}},
	}
	for _, d := range testData {
		config := Mapper{}
		config.MergeMaps(&d.mainMap, &d.additionalMap)
		if !reflect.DeepEqual(d.expected, d.mainMap) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, d.mainMap)
		}
	}
}

// global map only used for the TestRetag function literal
var testRetagUpdateTags = map[string]string{}

func setTagTestFctSuccess(res *string, tag *TagItem) error {
	testRetagUpdateTags[tag.Name] = tag.Value
	return nil
}
func setTagTestFctFailure(res *string, tag *TagItem) error {
	return errors.New("Badaboom")
}

func TestRetag(t *testing.T) {
	configWorking := Mapper{
		KeyMap: []*KeyMapper{
			{KeyPattern: ".*production.*", Destination: []*TagItem{{Name: "Env", Value: "prd"}}},
			{KeyPattern: ".*staging.*", Destination: []*TagItem{{Name: "Env", Value: "stg"}}},
			{KeyPattern: ".*apache.*", Destination: []*TagItem{{Name: "Team", Value: "web"}, {Name: "Component", Value: "apache"}}},
		},
		TagMap: []*TagMapper{
			{Source: &TagItem{Name: "Name", Value: ".*prod.*"}, Destination: []*TagItem{{Name: "Env", Value: "prd"}}},
			{Source: &TagItem{Name: "Name", Value: ".*dev.*"}, Destination: []*TagItem{{Name: "Env", Value: "dev"}}},
			{Source: &TagItem{Name: "Name", Value: ".*dev.*data.*"}, Destination: []*TagItem{{Name: "Team", Value: "data"}, {Name: "Env", Value: "dev"}}},
		},
		CopyTag: []*TagCopy{{Source: []string{"ENVIRONMENT", "ENVIRONMETNT", "Account"}, Destination: "Env"}},
		Sanity: []*TagSanity{
			{TagName: "Env", Transform: map[string][]string{"prd": {"prod", "production", "global"}, "stg": {"staging"}, "dev": {"development"}}},
			{TagName: "Team", Transform: map[string][]string{"web": {}}},
		},
		DefaultTagValues: map[string]string{"Env": "unknown", "Team": "unknown", "Service": "unknown"},
	}

	testData := []struct {
		resourceID    string
		tags          map[string]string
		keys          []string
		setTags       putTagFn
		logEntries    int
		expected      map[string]string
		config        Mapper
		expectedError error
	}{
		// empty source tag and keys
		{"my resource", map[string]string{}, []string{}, setTagTestFctSuccess, 3, map[string]string{"Env": "unknown", "Team": "unknown", "Service": "unknown"}, configWorking, nil},
		// non-matching tag existence, empty key
		{"my resource", map[string]string{"foo": "bar"}, []string{}, setTagTestFctSuccess, 4, map[string]string{"Env": "unknown", "Team": "unknown", "Service": "unknown"}, configWorking, nil},
		// 1 matching tag existence with transformation, empty key
		{"my resource", map[string]string{"Env": "prod", "Service": "whatever"}, []string{}, setTagTestFctSuccess, 2, map[string]string{"Env": "prd", "Team": "unknown"}, configWorking, nil},
		// 1 matching tag existence without transformation, empty key
		{"my resource", map[string]string{"Env": "prd", "Service": "whatever"}, []string{}, setTagTestFctSuccess, 2, map[string]string{"Team": "unknown"}, configWorking, nil},
		// 1 matching tag existence, matching key
		{"my resource", map[string]string{"Env": "prd", "Service": "whatever"}, []string{"web-apache"}, setTagTestFctSuccess, 2, map[string]string{"Team": "web", "Component": "apache"}, configWorking, nil},
		// 1 matching tag existence with transformation, overlapping key
		{"my resource", map[string]string{"Env": "prod", "Service": "whatever"}, []string{"non-staging-stuff"}, setTagTestFctSuccess, 2, map[string]string{"Env": "prd", "Team": "unknown"}, configWorking, nil},
		// 1 matching copy tag with transformation, overlapping key
		{"my resource", map[string]string{"Account": "prod", "Service": "whatever"}, []string{"non-staging-stuff"}, setTagTestFctSuccess, 3, map[string]string{"Env": "prd", "Team": "unknown"}, configWorking, nil},
		// 1 matching copy tag, partially overlapping tag map, partially overlapping key
		{"my resource", map[string]string{"Account": "prd", "Name": "dev-data-app", "Service": "Alice in chains", "noiseTag": "blah"}, []string{"non-staging-apache"}, setTagTestFctSuccess, 7, map[string]string{"Env": "prd", "Team": "data", "Component": "apache"}, configWorking, nil},
		// setTag errors out
		{"my resource", map[string]string{"Env": "prd", "Service": "whatever"}, []string{}, setTagTestFctFailure, 3, map[string]string{}, configWorking, errors.New("Failed to set tag on resource")},
		// bad config errors out
		{"my resource", map[string]string{"Name": "prod", "Service": "whatever"}, []string{}, setTagTestFctSuccess, 3, map[string]string{}, Mapper{CopyTag: []*TagCopy{{Source: []string{"Accou)nt"}, Destination: "Env"}}}, errors.New("GetFromTags failed")},
		{"my resource", map[string]string{"Service": "whatever"}, []string{"bla"}, setTagTestFctSuccess, 2, map[string]string{}, Mapper{KeyMap: []*KeyMapper{{KeyPattern: ".*a)b.*", Destination: []*TagItem{{Name: "Env", Value: "prd"}}}}}, errors.New("GetFromKey failed")},
		{"my resource", map[string]string{"Env": "prd", "Service": "whatever"}, []string{}, setTagTestFctSuccess, 2, map[string]string{}, Mapper{Sanity: []*TagSanity{{TagName: "Service", Transform: map[string][]string{"web": {"a)b"}}}}}, errors.New("ValidateTag failed")},
	}

	logger, hook := logrus_test.NewNullLogger()
	log = logrus.NewEntry(logger)

	for _, d := range testData {
		// Reset the counters
		hook.Reset()
		testRetagUpdateTags = map[string]string{}

		d.config.Retag(&d.resourceID, &d.tags, d.keys, d.setTags)
		if !reflect.DeepEqual(d.expected, testRetagUpdateTags) {
			t.Errorf("Expecting: %v\nGot: %v\n", d.expected, testRetagUpdateTags)
		}

		// Check what has been logged
		gotError := false
		for _, entry := range hook.Entries {
			if d.expectedError != nil && entry.Level == logrus.ErrorLevel {
				gotError = true
				if entry.Message != d.expectedError.Error() {
					t.Errorf("Unexpected message logged: got %s expecting %s for test case: %v\n", d.expectedError.Error(), entry.Message, d)
				}
			} else {
				if entry.Message != "Sanity check failed" && entry.Message != "No sanity configuration found" {
					t.Errorf("Unexpected message logged: %s for test case: %v\n", entry.Message, d)
				}
			}
		}
		if d.expectedError != nil && !gotError {
			t.Errorf("Did not get expected error: %s for test case: %v\n", d.expectedError.Error(), d)
		}
		if len(hook.Entries) != d.logEntries {
			t.Errorf("Unexpected number of messages logged. Got %d, expecting %d\nFor test case: %v\n", len(hook.Entries), d.logEntries, d)
		}
	}
}
