package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/sirupsen/logrus"
)

// CloudFrontProcessor holds the cloudfront-related actions
type CloudFrontProcessor struct {
	svc cloudfrontiface.CloudFrontAPI
}

// NewCloudFrontProcessor creates a new instance of CloudFrontProcessor containing an already
// initialized cloudfront client
func NewCloudFrontProcessor(sess *session.Session) *CloudFrontProcessor {
	return &CloudFrontProcessor{svc: cloudfront.New(sess)}
}

// TagsToMap transform the cloudfront tags structure into a map[string]string for
// easier manipulations
func (p *CloudFrontProcessor) TagsToMap(tagsInput []*cloudfront.Tag) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*tag.Key] = *tag.Value
	}
	return tagsHash
}

// SetTags sets tags on an cloudfront resource
func (p *CloudFrontProcessor) SetTags(resourceID *string, tags []*TagItem) error {
	newTags := []*cloudfront.Tag{}
	for _, tag := range tags {
		newTags = append(newTags, &cloudfront.Tag{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)})
	}

	input := &cloudfront.TagResourceInput{
		Resource: resourceID,
		Tags:     &cloudfront.Tags{Items: newTags},
	}
	_, err := p.svc.TagResource(input)
	return err
}

// GetTags gets the tags allocated to an cloudfront resource
func (p *CloudFrontProcessor) GetTags(resourceID *string) ([]*cloudfront.Tag, error) {
	input := &cloudfront.ListTagsForResourceInput{
		Resource: resourceID,
	}

	result, err := p.svc.ListTagsForResource(input)
	var tagsResult []*cloudfront.Tag
	if result.Tags != nil {
		tagsResult = (*result.Tags).Items
	}
	return tagsResult, err
}

// RetagDistributions parses all distributions and retags them
func (p *CloudFrontProcessor) RetagDistributions(m *Mapper) {
	err := p.svc.ListDistributionsPages(&cloudfront.ListDistributionsInput{},
		func(page *cloudfront.ListDistributionsOutput, lastPage bool) bool {
			if page.DistributionList != nil {
				for _, dist := range (*page.DistributionList).Items {
					t, err := p.GetTags(dist.ARN)
					if err != nil {
						log.WithFields(logrus.Fields{"error": err, "resource": *dist.ARN}).Fatal("Failed to get CloudFront distribution tags")
					}

					tags := p.TagsToMap(t)
					keys := []string{}
					if dist.Id != nil {
						keys = append(keys, *dist.Id)
					}
					if dist.DomainName != nil {
						keys = append(keys, *dist.DomainName)
					}
					if dist.Origins != nil {
						for _, orig := range (*dist.Origins).Items {
							keys = append(keys, *orig.DomainName)
						}
					}
					if dist.Aliases != nil {
						for _, alias := range (*dist.Aliases).Items {
							keys = append(keys, *alias)
						}
					}
					if dist.Comment != nil {
						keys = append(keys, *dist.Comment)
					}
					m.Retag(dist.ARN, &tags, keys, p.SetTags)
				}
			}
			return !lastPage
		})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("ListDistributionsPages failed")
	}
}
