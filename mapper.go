package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"regexp"
)

// TagItem is a standard AWS tag structure
type TagItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TagMapper makes the relation between an existing tag on a resource and a list
// of tags that should be present on that resource
type TagMapper struct {
	Source      *TagItem   `json:"source"`
	Destination []*TagItem `json:"destination"`
}

// KeyMapper makes the relation between an existing key and a list of
// tags that should be present on that resource. A key is:
// - the SSH Key name for an ec2 instance
// - DBClusterIdentifier, DBInstanceIdentifier, DBName, MasterUsername for RDS instances
// - DBClusterIdentifier, DBName, MasterUsername for RDS clusters
type KeyMapper struct {
	KeyPattern  string     `json:"pattern"`
	Destination []*TagItem `json:"destination"`
}

// TagCopy specify a list of tags you want to copy the value from if they exist
// before the sanity of the tag is processed
type TagCopy struct {
	Source      []string `json:"sources"`
	Destination string   `json:"destination"`
}

// TagSanity limits the values of a tag to a list of values after remapping the
// values using the Transform matrix (the key is the acceptable value and the
// values are the list of values it will be transformed from)
type TagSanity struct {
	TagName   string              `json:"tag_name"`
	Transform map[string][]string `json:"remap"`
}

// Mapper contains the different mappings between attributes and the list of
// tags that should be present on that resource
type Mapper struct {
	CopyTag          []*TagCopy        `json:"copy_tags,omitempty"`
	TagMap           []*TagMapper      `json:"tags,omitempty"`
	KeyMap           []*KeyMapper      `json:"keys,omitempty"`
	Sanity           []*TagSanity      `json:"sanity,omitempty"`
	DefaultTagValues map[string]string `json:"defaults,omitempty"`
}

// LoadConfig loads a json-formatted config into the current Mapper using the
// given io.Reader
func (m *Mapper) LoadConfig(configReader io.Reader) error {
	jsonParser := json.NewDecoder(configReader)
	return jsonParser.Decode(m)
}

// StripDefaults removes from the existing tags the ones that are set to the
// default value
func (m *Mapper) StripDefaults(existingTags *map[string]string) {
	for dKey, dVal := range m.DefaultTagValues {
		if val, ok := (*existingTags)[dKey]; ok {
			if val == dVal {
				delete(*existingTags, dKey)
			}
		}
	}
}

// GetMissingDefaults returns the tags and their default values for the ones that are
// missing from the existing tags
func (m *Mapper) GetMissingDefaults(existingTags *map[string]string) *map[string]string {
	result := make(map[string]string)
	for dKey, dVal := range m.DefaultTagValues {
		if _, ok := (*existingTags)[dKey]; !ok {
			result[dKey] = dVal
		}
	}
	return &result
}

// GetFromTags returns all the tags to add or update based on a list of existing
// tag that exist on the resource.
func (m *Mapper) GetFromTags(existingTags *map[string]string) (*map[string]string, error) {
	result := make(map[string]string)
	tagCp, err := m.getFromTagCopy(existingTags)
	if err != nil {
		return nil, err
	}
	result = *tagCp

	tagM, err := m.getFromTagMap(existingTags)
	if err != nil {
		return nil, err
	}
	// Because we update existingTags map at the same time, we don't need to worry
	// about overwritten values
	for k, v := range *tagM {
		result[k] = v
	}
	return &result, nil
}

// getFromTagCopy returns the tags matching from the TagCopy mapping
func (m *Mapper) getFromTagCopy(existingTags *map[string]string) (*map[string]string, error) {
	var (
		match bool
		err   error
	)
	result := make(map[string]string)
	// The order of the tags defined in Source matters and the tag names are
	// regex-based with case-insensitive so we can't just check if the keys of
	// existing tags exist inside the sources array
	for _, tagCp := range m.CopyTag {
		// Skip is the destination tag is already set
		if _, ok := (*existingTags)[tagCp.Destination]; ok {
			continue
		}
	CopyTagSourcesLoop:
		for _, src := range tagCp.Source {
			for k, v := range *existingTags {
				if match, err = regexp.MatchString("(?i)^"+src+"$", k); err != nil {
					return nil, err
				}
				if match {
					result[tagCp.Destination] = v
					// Also register as existing tag for easier use in the rest of the
					// functions
					(*existingTags)[tagCp.Destination] = v
					// Only take the 1st matching source
					break CopyTagSourcesLoop
				}
			}
		}
	}
	return &result, err
}

// getFromTagMap returns the tags to create/update based on the current TagMap
// configuration. The name of the tag in the source is case-sensitive and not
// parsed with a regex. Its value is parsed with a case-insensitive regex
func (m *Mapper) getFromTagMap(existingTags *map[string]string) (*map[string]string, error) {
	var (
		match bool
		err   error
	)
	result := make(map[string]string)
	for _, mapping := range m.TagMap {
		if val, ok := (*existingTags)[mapping.Source.Name]; ok {
			if match, err = regexp.MatchString("(?i)^"+mapping.Source.Value+"$", val); err != nil {
				return nil, err
			}
			if match {
				for _, dst := range mapping.Destination {
					if _, ok := (*existingTags)[dst.Name]; ok {
						continue // skip if tag already set
					}
					result[dst.Name] = dst.Value
					// Also register as existing tag for easier use in the rest of the
					// functions
					(*existingTags)[dst.Name] = dst.Value
				}
			}
		}
	}
	return &result, err
}

// GetFromKey retrieves the tags corresponding to the KeyMap configuration
// except when the tag is already set in existingTags
func (m *Mapper) GetFromKey(resourceKey string, existingTags *map[string]string) (*map[string]string, error) {
	var (
		match bool
		err   error
	)
	result := make(map[string]string)
	for _, keyM := range m.KeyMap {
		if match, err = regexp.MatchString("(?i)^"+keyM.KeyPattern+"$", resourceKey); err != nil {
			return nil, err
		}
		if match {
			for _, dst := range keyM.Destination {
				if _, ok := (*existingTags)[dst.Name]; ok {
					continue // skip if tag already in existingTags
				}
				if _, ok := result[dst.Name]; ok {
					continue // skip if tag already set
				}
				result[dst.Name] = dst.Value
			}
		}
	}
	return &result, err
}

// ValidateTag operates on the tags map to validates a given tag
// Sanity configuration element of the Mapper
func (m *Mapper) ValidateTag(tagName, tagValue string) (*TagItem, error) {
	var (
		match bool
		err   error
	)

	result := TagItem{Name: tagName, Value: tagValue}
	for _, elt := range m.Sanity {
		if elt.TagName != tagName {
			continue
		}
		for ref, alt := range elt.Transform {
			if tagValue == ref {
				return &result, nil // avoid some costly regexp if the current value is already clean
			}
			for _, val := range alt {
				if match, err = regexp.MatchString("(?i)^"+val+"$", tagValue); err != nil {
					return &result, err
				}
				if match {
					result.Value = ref
					return &result, nil
				}
			}
		}
		return &result, NewErrSanityNoMapping("No match found for the sanity check", tagName, tagValue)
	}
	return &result, NewErrSanityConfig("No sanity configuration found", tagName)
}

// MergeMaps adds to mainMap the missing elements that are present in
// complementary
func (m *Mapper) MergeMaps(mainMap, complementary *map[string]string) {
	for k, v := range *complementary {
		if _, ok := (*mainMap)[k]; !ok {
			(*mainMap)[k] = v
		}
	}
}

type putTagFn func(*string, *TagItem) error

// Retag does the different re-tagging operations and calls the given setTag function
func (m *Mapper) Retag(resourceID *string, tags *map[string]string, keys []string, setTag putTagFn) {
	var (
		newTags, mapFromKey, mapFromMissing *map[string]string
		err                                 error
	)
	m.StripDefaults(tags)
	if newTags, err = m.GetFromTags(tags); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("GetFromTags failed")
	}

	for _, item := range keys {
		if mapFromKey, err = m.GetFromKey(item, tags); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Error("GetFromKey failed")
		}
		m.MergeMaps(newTags, mapFromKey)
	}
	mapFromMissing = m.GetMissingDefaults(tags)
	m.MergeMaps(newTags, mapFromMissing)

	// This part evaluates if the existing tags need to be updated
	sanitized := make(map[string]string)
	for k, v := range *tags {
		sanitizedTag, err := m.ValidateTag(k, v)
		if err != nil {
			switch err.(type) {
			case *ErrSanityNoMapping:
				subErr, _ := err.(*ErrSanityNoMapping)
				log.WithFields(logrus.Fields{"error": subErr, "resource": *resourceID, "tag_name": subErr.TagName, "tag_value": subErr.TagValue}).Info("Sanity check failed")
			case *ErrSanityConfig:
				subErr, _ := err.(*ErrSanityConfig)
				log.WithFields(logrus.Fields{"error": subErr, "resource": *resourceID, "tag_name": subErr.TagName}).Warn("Sanity check failed")
			default:
				log.WithFields(logrus.Fields{"error": err}).Error("ValidateTag failed")
			}
		}
		if sanitizedTag != nil {
			if sanitizedTag.Value != v {
				sanitized[k] = sanitizedTag.Value
			}
		}
	}
	m.MergeMaps(newTags, &sanitized)

	for k, v := range *newTags {
		finalTag, err := m.ValidateTag(k, v)
		if err != nil {
			switch err.(type) {
			case *ErrSanityNoMapping:
				subErr, _ := err.(*ErrSanityNoMapping)
				log.WithFields(logrus.Fields{"error": subErr, "resource": *resourceID, "tag_name": subErr.TagName, "tag_value": subErr.TagValue}).Info("Sanity check failed")
			case *ErrSanityConfig:
				subErr, _ := err.(*ErrSanityConfig)
				log.WithFields(logrus.Fields{"error": subErr, "resource": *resourceID, "tag_name": subErr.TagName}).Warn("Sanity check failed")
			default:
				log.WithFields(logrus.Fields{"error": err}).Error("ValidateTag failed")
			}
		}
		log.WithFields(logrus.Fields{"resource": *resourceID, "tag_name": (*finalTag).Name, "tag_value": (*finalTag).Value}).Debug("Setting tag on resource")
		if err = setTag(resourceID, finalTag); err != nil {
			log.WithFields(logrus.Fields{"error": err, "resource": *resourceID}).Error("Failed to set tag on resource")
		}
	}
}
