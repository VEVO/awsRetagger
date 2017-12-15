package providers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"

	"github.com/VEVO/awsRetagger/mapper"
)

// S3Processor holds the s3-related actions
type S3Processor struct {
	svc    s3iface.S3API
	region *string
}

// NewS3Processor creates a new instance of S3Processor containing an already
// initialized s3 client
func NewS3Processor(sess *session.Session) *S3Processor {
	return &S3Processor{svc: s3.New(sess), region: sess.Config.Region}
}

// TagsToMap transform the s3 tags structure into a map[string]string for
// easier manipulations
func (e *S3Processor) TagsToMap(tagsInput []*s3.Tag) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*tag.Key] = *tag.Value
	}
	return tagsHash
}

// SetTags sets tags on a s3 bucket
func (e *S3Processor) SetTags(resourceID *string, tags []*mapper.TagItem) error {
	newTags := s3.Tagging{}
	for _, tag := range tags {
		if len((*tag).Name) > 0 {
			newTags.TagSet = append(newTags.TagSet, &s3.Tag{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)})
		}
	}

	if len(newTags.TagSet) == 0 {
		return nil
	}
	_, err := e.svc.PutBucketTagging(&s3.PutBucketTaggingInput{Bucket: resourceID, Tagging: &newTags})
	return err
}

// RetagBuckets parses all buckets and retags them
func (e *S3Processor) RetagBuckets(m mapper.Iface) {
	result, err := e.svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("ListBuckets failed")
	}

	for _, bucket := range result.Buckets {
		// Makes sure we are in the right region and avoid stuffs like:
		// AuthorizationHeaderMalformed: The authorization header is malformed; the region 'us-east-1' is wrong
		if location, err := e.svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: bucket.Name}); err != nil {
			log.WithFields(logrus.Fields{"bucket": *bucket.Name, "error": err.Error()}).Fatal("GetBucketLocation failed")
		} else {
			loc := ""
			if location.LocationConstraint != nil {
				loc = *location.LocationConstraint
			}
			loc = s3.NormalizeBucketLocation(loc)
			if loc != *e.region {
				log.WithFields(logrus.Fields{"bucket": *bucket.Name, "location": loc}).Debug("Skipping bucket in different region than session")
				continue // skip if the bucket is not the one specified in the session region
			}
		}
		// Get the tags
		bTags, err := e.svc.GetBucketTagging(&s3.GetBucketTaggingInput{Bucket: bucket.Name})
		if err != nil {
			emptyTag := false
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case "NoSuchTagSet":
					emptyTag = true // ignore errors when not tagset associated to the bucket
				}
			}
			if !emptyTag {
				log.WithFields(logrus.Fields{"bucket": *bucket.Name, "error": err.Error()}).Fatal("GetBucketTagging failed")
			}
		}
		tags := e.TagsToMap(bTags.TagSet)
		keys := []string{}
		if bucket.Name != nil {
			keys = append(keys, *bucket.Name)
		}
		m.Retag(bucket.Name, &tags, keys, e.SetTags)
	}
}
