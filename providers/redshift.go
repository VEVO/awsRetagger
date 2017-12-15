package providers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/redshift/redshiftiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"

	"github.com/VEVO/awsRetagger/mapper"
)

// RedshiftProcessor holds the redshift-related actions
type RedshiftProcessor struct {
	svc       redshiftiface.RedshiftAPI
	region    *string
	accountID *string
}

// NewRedshiftProcessor creates a new instance of RedshiftProcessor containing an already
// initialized redshift client
func NewRedshiftProcessor(sess *session.Session) (*RedshiftProcessor, error) {
	stsSvc := sts.New(sess)
	input := &sts.GetCallerIdentityInput{}
	accInfo, err := stsSvc.GetCallerIdentity(input)
	if err != nil {
		return nil, err
	}
	return &RedshiftProcessor{svc: redshift.New(sess), region: (*sess.Config).Region, accountID: accInfo.Account}, nil
}

// TagsToMap transform the redshift tags structure into a map[string]string for
// easier manipulations
func (p *RedshiftProcessor) TagsToMap(tagsInput []*redshift.TaggedResource) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*(*tag.Tag).Key] = *(*tag.Tag).Value
	}
	return tagsHash
}

// SetTags sets tags on an redshift resource
func (p *RedshiftProcessor) SetTags(resourceID *string, tags []*mapper.TagItem) error {
	newTags := []*redshift.Tag{}
	for _, tag := range tags {
		newTags = append(newTags, &redshift.Tag{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)})
	}

	input := &redshift.CreateTagsInput{
		ResourceName: resourceID,
		Tags:         newTags,
	}
	_, err := p.svc.CreateTags(input)
	return err
}

// GetTags gets the tags allocated to an redshift resource
func (p *RedshiftProcessor) GetTags(resourceID *string) ([]*redshift.TaggedResource, error) {
	input := &redshift.DescribeTagsInput{
		ResourceName: resourceID,
	}

	result, err := p.svc.DescribeTags(input)
	return result.TaggedResources, err
}

// getArn builds the arn for the given resourceIdentifier:
// http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-redshift
func (p *RedshiftProcessor) getArn(resourceType, resourceIdentifier string) string {
	res := arn.ARN{Partition: "aws", Service: "redshift", Region: *p.region, AccountID: *p.accountID, Resource: resourceType + ":" + resourceIdentifier}
	return res.String()
}

// RetagClusters parses all clusters and retags them
func (p *RedshiftProcessor) RetagClusters(m *mapper.Mapper) {
	err := p.svc.DescribeClustersPages(&redshift.DescribeClustersInput{},
		func(page *redshift.DescribeClustersOutput, lastPage bool) bool {
			for _, elt := range page.Clusters {
				clArn := p.getArn("cluster", *elt.ClusterIdentifier)
				t, err := p.GetTags(&clArn)
				if err != nil {
					log.WithFields(logrus.Fields{"error": err, "resource": *elt.ClusterIdentifier}).Fatal("Failed to get Redshift cluster tags")
				}
				tags := p.TagsToMap(t)
				keys := []string{}
				if elt.ClusterIdentifier != nil {
					keys = append(keys, *elt.ClusterIdentifier)
				}
				if elt.DBName != nil {
					keys = append(keys, *elt.DBName)
				}
				if elt.MasterUsername != nil {
					keys = append(keys, *elt.MasterUsername)
				}
				m.Retag(&clArn, &tags, keys, p.SetTags)
			}
			return !lastPage
		})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("DescribeClusters failed")
	}
}
