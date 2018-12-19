package search

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	bugsnag "github.com/bugsnag/bugsnag-go"
)

// This is 17 characters plus the "i-" prefix
const ExactEc2InstanceIDLength = 19
const EnvVarPrefixForAwsRoles = "AWS_ROLE_"

type ec2SDK interface {
	DescribeInstancesWithContext(ctx aws.Context, input *ec2.DescribeInstancesInput, opts ...request.Option) (*ec2.DescribeInstancesOutput, error)
}

// buildEc2ClientsFromEnvironment uses environment variables to build instances
// of the EC2 client library for each AWS account it should discover resources
// within.
//
// The main variables for configuration are:
//
// `AWS_ROLE_{account alias}` - The role slash-infra should assume to gain access
// to the account known as {account alias}
//
// `AWS_REGION_{account alias}` - If the account's resources are in a region
// other than us-east-1, specify it here.
//
// If an account uses several regions, then you can specify role several times
// under different aliases. e.g.
//
// ```
// AWS_ROLE_DEV_US_EAST=...
// AWS_ROLE_DEV_EU=...
// ```
func buildEc2ClientsFromEnvironment() []ec2SDK {
	clients := []ec2SDK{}
	environ := os.Environ()

	for _, pair := range environ {
		if !strings.HasPrefix(pair, EnvVarPrefixForAwsRoles) {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		roleArn := parts[1]

		awsAccountAlias := key[len(EnvVarPrefixForAwsRoles):]

		// Some of our infra is not in us-east-1 (e.g. dev-vpc)
		// Allow slash-infra to create clients that will discover resources in those regions
		region := os.Getenv(fmt.Sprintf("AWS_REGION_%s", awsAccountAlias))
		if region == "" {
			region = "us-east-1"
		}

		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewEnvCredentials(),
			// Setting here rather than in env variables as not all of our
			// accounts are in us-east-1
			Region: aws.String(region),
		}))
		creds := stscreds.NewCredentials(sess, roleArn)
		svc := ec2.New(sess, &aws.Config{Credentials: creds})

		clients = append(clients, svc)
	}

	return clients
}

func NewEc2() *EC2Resolver {
	return &EC2Resolver{clients: buildEc2ClientsFromEnvironment()}
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
