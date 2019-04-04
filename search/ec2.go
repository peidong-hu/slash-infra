package search

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	bugsnag "github.com/bugsnag/bugsnag-go"
)

// This is 17 characters plus the "i-" prefix
const ExactEc2InstanceIDLength = 19

type ec2SDK interface {
	DescribeInstancesWithContext(ctx aws.Context, input *ec2.DescribeInstancesInput, opts ...request.Option) (*ec2.DescribeInstancesOutput, error)
}

func NewEc2(awsSessions map[string]*session.Session) *EC2Resolver {
	clients := make([]ec2SDK, 0)

	for _, sess := range awsSessions {
		clients = append(clients, ec2.New(sess))
	}

	return &EC2Resolver{clients: clients}
}

type Result struct {
	Kind     string
	Metadata map[string][]string
	Links    map[string]string
}

func (r Result) GetMetadata(key string) string {
	set, ok := r.Metadata[key]
	if !ok {
		return ""
	}

	return strings.Join(set, ", ")
}

func (r Result) GetLink(key string) string {
	url, ok := r.Links[key]
	if !ok {
		return ""
	}

	return url
}

type ResultSet struct {
	Kind       string
	SearchLink string
	Results    []Result
}

type EC2Resolver struct {
	clients []ec2SDK
}

func (e *EC2Resolver) Search(ctx context.Context, query string) []ResultSet {
	results := []ResultSet{}

	query = strings.TrimSpace(query)

	for _, client := range e.clients {
		result, err := findEC2InstancesByID(ctx, client, query)

		if err != nil {
			log.Print(err)
		}

		if result != nil {
			results = append(results, *result)
		}

	}

	return results
}

func findEC2InstancesByID(ctx context.Context, client ec2SDK, search string) (*ResultSet, error) {
	// EC2 instance IDs have a very specific format
	if !strings.HasPrefix(search, "i-") {
		return nil, nil
	}

	// The EC2 API does not allow you to do substring searches
	if len(search) != ExactEc2InstanceIDLength {
		return nil, nil
	}

	output, err := client.DescribeInstancesWithContext(
		ctx,
		&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				&ec2.Filter{Name: aws.String("instance-id"), Values: []*string{aws.String(search)}},
			},
		},
	)

	if err != nil {
		bugsnag.Notify(err)
		return nil, err
	}

	results := []Result{}

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			publicIpAddresses := []string{}
			privateIpAddresses := []string{}

			// Stopped instances do not appear to have network interfaces
			if instance.NetworkInterfaces != nil {
				for _, networkInterface := range instance.NetworkInterfaces {
					if networkInterface == nil {
						continue
					}

					if networkInterface.Association != nil {
						publicIpAddresses = append(publicIpAddresses, *networkInterface.Association.PublicIp)
					}

					if networkInterface.PrivateIpAddresses != nil {
						for _, privateIp := range networkInterface.PrivateIpAddresses {
							privateIpAddresses = append(privateIpAddresses, *privateIp.PrivateIpAddress)
						}
					}
				}
			}

			result := Result{
				Kind: "ec2.instance",
				Metadata: map[string][]string{
					"instance_id":    []string{*instance.InstanceId},
					"ami_id":         []string{*instance.ImageId},
					"instance_type":  []string{*instance.InstanceType},
					"instance_state": []string{*instance.State.Name},
					"az":             []string{*instance.Placement.AvailabilityZone},
					"public_ips":     publicIpAddresses,
					"private_ips":    privateIpAddresses,
				},
				Links: map[string]string{
					"ec2_console":     ec2ConsoleLink("us-east-1", *instance.InstanceId),
					"config_timeline": ec2ConfigTimelineLink("us-east-1", *instance.InstanceId),
				},
			}

			for _, tag := range instance.Tags {
				result.Metadata[fmt.Sprintf("tag:%s", *tag.Key)] = []string{*tag.Value}
			}

			results = append(results, result)
		}
	}

	return &ResultSet{Kind: "ec2.instance", Results: results}, err
}

func ec2ConsoleLink(region, search string) string {
	return fmt.Sprintf("https://console.aws.amazon.com/ec2/v2/home?region=%s#Instances:search=%s;sort=desc:launchTime", region, search)
}

func ec2ConfigTimelineLink(region, instanceId string) string {
	return fmt.Sprintf("https://console.aws.amazon.com/config/home?region=%s#/timeline/AWS::EC2::Instance/%s/configuration", region, instanceId)
}
