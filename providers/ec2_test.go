package providers

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/VEVO/awsRetagger/mapper"
)

// mockEc2Client is ised to mock calls to the ec2 API
type mockEc2Client struct {
	ec2iface.EC2API
	// ResourceID is the resource that has been passed to the mocked function
	ResourceIDs []*string
	// ResourceTags are the tags that have been passed to the mocked function when
	// setting or that is available on the mocked resource when getting
	ResourceTags []*ec2.Tag
	// ReturnError is the error that you want your mocked function to return
	ReturnError error
}

func (m *mockEc2Client) CreateTags(input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	m.ResourceIDs = append(m.ResourceIDs, input.Resources...)
	if len(input.Tags) > 0 {
		m.ResourceTags = append(m.ResourceTags, input.Tags...)
	}
	return &ec2.CreateTagsOutput{}, m.ReturnError

}

func TestEc2TagsToMap(t *testing.T) {
	testData := []struct {
		inputTags  []*ec2.Tag
		outputTags map[string]string
	}{
		{[]*ec2.Tag{}, map[string]string{}},
		{[]*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, map[string]string{"foo": "bar"}},
		{[]*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, map[string]string{"foo": "bar", "Aerosmith": "rocks"}},
	}
	for _, d := range testData {
		p := Ec2Processor{}

		res := p.TagsToMap(d.inputTags)
		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}

func TestEc2SetTags(t *testing.T) {
	testData := []struct {
		inputResource           string
		outputResource          []*string
		inputTags               []*mapper.TagItem
		outputTags              []*ec2.Tag
		inputError, outputError error
	}{
		{"my resource", []*string{}, []*mapper.TagItem{{}}, []*ec2.Tag{}, nil, nil},
		{"my resource", []*string{aws.String("my resource")}, []*mapper.TagItem{{Name: "foo", Value: "bar"}}, []*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, nil, nil},
		{"my resource", []*string{aws.String("my resource")}, []*mapper.TagItem{{Name: "foo", Value: "bar"}, {Name: "Aerosmith", Value: "rocks"}}, []*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, nil, nil},
		{"my resource", []*string{aws.String("my resource")}, []*mapper.TagItem{{Name: "foo", Value: "bar"}}, []*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, errors.New("Badaboom"), errors.New("Badaboom")},
	}
	for _, d := range testData {
		mockSvc := &mockEc2Client{ResourceIDs: []*string{}, ReturnError: d.inputError, ResourceTags: []*ec2.Tag{}}
		p := Ec2Processor{svc: mockSvc}

		err := p.SetTags(&d.inputResource, d.inputTags)
		if !reflect.DeepEqual(err, d.outputError) {
			t.Errorf("Expecting error: %v\nGot: %v\n", d.outputError, err)
		}

		if len(mockSvc.ResourceIDs) != len(d.outputResource) || (len(mockSvc.ResourceIDs) > 0 && *mockSvc.ResourceIDs[0] != *d.outputResource[0]) {
			t.Errorf("Expecting to update resource: %v, got: %v\n", d.outputResource, mockSvc.ResourceIDs)
		}

		if !reflect.DeepEqual(mockSvc.ResourceTags, d.outputTags) {
			t.Errorf("Expecting to update tag: %v\nGot: %v\n", d.outputTags, mockSvc.ResourceTags)
		}
	}
}
