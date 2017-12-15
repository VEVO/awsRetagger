package providers

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"
	logrus_test "github.com/sirupsen/logrus/hooks/test"

	"github.com/VEVO/awsRetagger/mapper"
)

// mockS3Client is used to mock s3 calls
type mockS3Client struct {
	s3iface.S3API
	// ResourceID is the resource that has been passed to the mocked function
	ResourceID *string
	// ResourceTags are the tags that have been passed to the mocked function when
	// setting or that is available on the mocked resource when getting
	ResourceTags []*s3.Tag
	// ReturnError is the error that you want your mocked function to return
	ReturnError error
	// BucketsNRegions is the list of buckets that ListBuckets pulls its output
	// from as keys and regions as value
	BucketsNRegions map[string]string
	// BucketsTags is the list of tags corresponding to the buckets for
	// GetBucketTagging
	BucketsTags map[string][]*s3.Tag
	// BucketsErrors output error of a given bucket for GetBucketTagging
	BucketsErrors map[string]error
}

func (m *mockS3Client) PutBucketTagging(input *s3.PutBucketTaggingInput) (*s3.PutBucketTaggingOutput, error) {
	m.ResourceID = input.Bucket
	if input.Tagging != nil {
		m.ResourceTags = append(m.ResourceTags, input.Tagging.TagSet...)
	}
	return &s3.PutBucketTaggingOutput{}, m.ReturnError
}

func (m *mockS3Client) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	buckets := []*s3.Bucket{}
	for bucket := range m.BucketsNRegions {
		buckets = append(buckets, &s3.Bucket{Name: aws.String(bucket)})
	}
	output := s3.ListBucketsOutput{Owner: &s3.Owner{}, Buckets: buckets}
	return &output, m.ReturnError
}

func (m *mockS3Client) GetBucketLocation(input *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	region, _ := m.BucketsNRegions[*input.Bucket]
	return &s3.GetBucketLocationOutput{LocationConstraint: aws.String(region)}, nil
}

func (m *mockS3Client) GetBucketTagging(input *s3.GetBucketTaggingInput) (*s3.GetBucketTaggingOutput, error) {
	var (
		tags      []*s3.Tag
		outputErr error
	)
	tags, _ = m.BucketsTags[*input.Bucket]
	outputErr, _ = m.BucketsErrors[*input.Bucket]
	return &s3.GetBucketTaggingOutput{TagSet: tags}, outputErr
}

func TestS3TagsToMap(t *testing.T) {
	testData := []struct {
		inputTags  []*s3.Tag
		outputTags map[string]string
	}{
		{[]*s3.Tag{}, map[string]string{}},
		{[]*s3.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, map[string]string{"foo": "bar"}},
		{[]*s3.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, map[string]string{"foo": "bar", "Aerosmith": "rocks"}},
	}
	for _, d := range testData {
		p := S3Processor{}

		res := p.TagsToMap(d.inputTags)
		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}

func TestS3SetTags(t *testing.T) {
	testData := []struct {
		inputResource, outputResource string
		inputTags                     []*mapper.TagItem
		outputTags                    []*s3.Tag
		inputError, outputError       error
	}{
		{"my resource", "", []*mapper.TagItem{{}}, []*s3.Tag{}, nil, nil},
		{"my resource", "my resource", []*mapper.TagItem{{Name: "foo", Value: "bar"}}, []*s3.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, nil, nil},
		{"my resource", "my resource", []*mapper.TagItem{{Name: "foo", Value: "bar"}, {Name: "Aerosmith", Value: "rocks"}}, []*s3.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, nil, nil},
		{"my resource", "my resource", []*mapper.TagItem{{Name: "foo", Value: "bar"}}, []*s3.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, errors.New("Badaboom"), errors.New("Badaboom")},
	}
	for _, d := range testData {
		mockSvc := &mockS3Client{ReturnError: d.inputError, ResourceTags: []*s3.Tag{}}
		p := S3Processor{svc: mockSvc}

		err := p.SetTags(&d.inputResource, d.inputTags)
		if !reflect.DeepEqual(err, d.outputError) {
			t.Errorf("Expecting error: %v\nGot: %v\n", d.outputError, err)
		}

		if mockSvc.ResourceID == nil {
			if d.outputResource != "" {
				t.Errorf("Expecting to update resource: %s, got no resourceID assigned\n", d.outputResource)
			}
		} else {
			if *mockSvc.ResourceID != d.outputResource {
				t.Errorf("Expecting to update resource: %s, got: %s\n", d.outputResource, *mockSvc.ResourceID)
			}
		}

		if !reflect.DeepEqual(mockSvc.ResourceTags, d.outputTags) {
			t.Errorf("Expecting to update tag: %v\nGot: %v\n", d.outputTags, mockSvc.ResourceTags)
		}
	}
}

func TestS3RetagBuckets(t *testing.T) {
	testData := []struct {
		sessionRegion        string
		inputBucketsNRegions map[string]string
		inputBucketsTags     map[string][]*s3.Tag
		outputBucketsTags    map[string]map[string]string
		outputBucketsKeys    map[string][]string
		inputBucketsErrors   map[string]error
	}{
		{"us-east-1", map[string]string{}, map[string][]*s3.Tag{}, nil, nil, map[string]error{}},
		{
			"us-east-1",
			map[string]string{"bucket1": "us-east-1", "bucket2": "us-east-1", "homerSimpson": "us-west-2"},
			map[string][]*s3.Tag{"bucket1": {&s3.Tag{Key: aws.String("Team"), Value: aws.String("Gryffindor")}, &s3.Tag{Key: aws.String("Strength"), Value: aws.String("chivalry")}}, "homerSimpson": {&s3.Tag{Key: aws.String("Team"), Value: aws.String("Nuclear")}}},
			map[string]map[string]string{"bucket1": {"Team": "Gryffindor", "Strength": "chivalry"}, "bucket2": {}},
			map[string][]string{"bucket1": {"bucket1"}, "bucket2": {"bucket2"}},
			map[string]error{},
		},
		{
			"us-west-2",
			map[string]string{"bucket1": "us-east-1", "bucket2": "us-east-1", "homerSimpson": "us-west-2"},
			map[string][]*s3.Tag{"bucket1": {&s3.Tag{Key: aws.String("Team"), Value: aws.String("Gryffindor")}, &s3.Tag{Key: aws.String("Strength"), Value: aws.String("chivalry")}}, "homerSimpson": {&s3.Tag{Key: aws.String("Team"), Value: aws.String("Nuclear")}}},
			map[string]map[string]string{"homerSimpson": {"Team": "Nuclear"}},
			map[string][]string{"homerSimpson": {"homerSimpson"}},
			map[string]error{},
		},
	}

	// silence the logs
	logger, _ := logrus_test.NewNullLogger()
	log = logrus.NewEntry(logger)

	for _, d := range testData {
		mockSvc := &mockS3Client{BucketsNRegions: d.inputBucketsNRegions, BucketsTags: d.inputBucketsTags, BucketsErrors: d.inputBucketsErrors}
		m := mapper.MockMapper{}
		p := S3Processor{svc: mockSvc, region: &d.sessionRegion}
		p.RetagBuckets(&m)

		if !reflect.DeepEqual(d.outputBucketsTags, m.ResourceTags) {
			t.Errorf("Expecting Mapper.Retag to receive tags: %v\nGot: %v\n", d.outputBucketsTags, m.ResourceTags)
		}

		if !reflect.DeepEqual(d.outputBucketsKeys, m.ResourceKeys) {
			t.Errorf("Expecting Mapper.Retag to receive keys: %v\nGot: %v\n", d.outputBucketsKeys, m.ResourceKeys)
		}

	}
}
