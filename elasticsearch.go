package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/sirupsen/logrus"
)

// ElkProcessor holds the elasticsearch-related actions
type ElkProcessor struct {
	svc *elasticsearchservice.ElasticsearchService
}

// NewElkProcessor creates a new instance of ElkProcessor containing an already
// initialized elasticsearchservice client
func NewElkProcessor(sess *session.Session) *ElkProcessor {
	return &ElkProcessor{svc: elasticsearchservice.New(sess)}
}

// TagsToMap transform the elasticsearchservice tags structure into a map[string]string for
// easier manipulations
func (p *ElkProcessor) TagsToMap(tagsInput []*elasticsearchservice.Tag) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*tag.Key] = *tag.Value
	}
	return tagsHash
}

// SetTag sets a tag on an elasticsearchservice resource
func (p *ElkProcessor) SetTag(resourceID *string, tag *TagItem) error {
	input := &elasticsearchservice.AddTagsInput{
		ARN:     resourceID,
		TagList: []*elasticsearchservice.Tag{{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)}},
	}
	_, err := p.svc.AddTags(input)
	return err
}

// GetTags gets the tags allocated to an elasticsearchservice resource
func (p *ElkProcessor) GetTags(resourceID *string) ([]*elasticsearchservice.Tag, error) {
	input := &elasticsearchservice.ListTagsInput{
		ARN: resourceID,
	}

	result, err := p.svc.ListTags(input)
	return result.TagList, err
}

// RetagDomains parses all elasticsearch domains and retags them
func (p *ElkProcessor) RetagDomains(m *Mapper) {
	result, err := p.svc.ListDomainNames(&elasticsearchservice.ListDomainNamesInput{})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("ListDomainNames failed")
	}

	for _, domain := range result.DomainNames {
		domInfo, err := p.svc.DescribeElasticsearchDomain(&elasticsearchservice.DescribeElasticsearchDomainInput{DomainName: domain.DomainName})
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "resource": *domain.DomainName}).Fatal("Failed to get Elasticsearch domain attributes")
		}
		dom := *(domInfo.DomainStatus)
		t, err := p.GetTags(dom.ARN)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "resource": *domain.DomainName}).Fatal("Failed to get Elasticsearch domain tags")
		}

		tags := p.TagsToMap(t)
		keys := []string{}
		if dom.DomainId != nil {
			keys = append(keys, *dom.DomainId)
		}
		if dom.DomainName != nil {
			keys = append(keys, *dom.DomainName)
		}
		m.Retag(dom.ARN, &tags, keys, p.SetTag)
	}
}
