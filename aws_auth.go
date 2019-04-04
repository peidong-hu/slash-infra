package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

const EnvVarPrefixForAwsRoles = "AWS_ROLE_"

// buildAwsSessionsFromEnvironment to generate AWS sessions that store
// credentials & config needed by the various AWS SDKs.
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
func buildAwsSessionsFromEnvironment() map[string]*session.Session {
	sessionsByAlias := make(map[string]*session.Session, 0)

	environ := os.Environ()

	// The credentials in this session only have permission to assume IAM
	// specific roles. It does not matter which account the credentials
	// belong to, so long as each of the roles we're using trusts it.
	assumeOnlySession := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	}))

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

		creds := stscreds.NewCredentials(assumeOnlySession, roleArn)

		roleSession := session.Must(session.NewSession(&aws.Config{
			Credentials: creds,
			Region:      aws.String(region),
		}))

		sessionsByAlias[strings.ToLower(awsAccountAlias)] = roleSession

	}

	return sessionsByAlias
}
