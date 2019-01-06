# slash-infra

This is a slack integration for "chatops". A lot of the hubot scripts
don't integrate natively with slack's slash commands, or give you
unnecessarily powerful commands (e.g. launching individual ec2 instances,
or creating auto-scaling groups), or only work against one AWS account.

This tool came from a need to lookup EC2 instances by their IP/instance
ID. Our production/staging/dev environments are hosted in different AWS
accounts, and some are hosted in different regions. Switching between
accounts and regions to track down an instance can be quite laborious
and error prone, especially if you're under pressure trying to triage a
problem.

`/infra-search {query}` can search multiple AWS accounts to find
resources. Currently it only supports looking up instances by their
instance ID.


## Configuring AWS access

Rather than create a user in each AWS account, the app uses a limited
IAM user to assume roles in each AWS account that should be searched.

This approach seems convoluted, but there are several benefits:

- When an IAM user "assumes" a role, AWS generates a set of temporary
  credentials for the user that have the same permissions as the role.
  These credentials are short-lived, and are rotated transparently by
  the AWS SDK. If these temporary credentials were to be leaked to a
  third party, they would only be usable for a short period of time.
- If the credentials for the IAM user are leaked, an attacker can't use
  the credentials unless they have the ARN of a role in a specific AWS
  account. This may not be a great comfort if the role resides in the
  same account as the user (you can use `aws sts get-caller-identity` to
  get the ID of the account the current credentials belong to), but
  there's nothing in `slash-infra` that requires a specific role name -
  you could choose a unique role name for your org, thus making it
  slightly more difficult for an attacker to exploit.

You can configure the IAM user using the conventional environment
variables:

```console
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

This IAM user should have the following permission policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowAssumingSlashInfraRoles",
            "Effect": "Allow",
            "Action": "sts:AssumeRole",
            "Resource": "arn:aws:iam::*:role/SlashInfraInspection"
        }
    ]
}
```

Note that the `*` in the ARN allows this suser to assume the
`SlashInfraInspection` role in any AWS account that:

- has that role
- has marked your AWS account ID as a "Trusted entity" in the role's
  "Trust relationships"

If these are new concepts for you, I'd really recommend reading [AWS'
documentation on IAM
roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_common-scenarios_aws-accounts.html)

Each role in each account should have a permission policy like this:


```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowReadOnlyAccess",
            "Effect": "Allow",
            "Action": "ec2:DescribeInstances",
            "Resource": "*"
        }
    ]
}
```

You can then configure `slash-infra` to use the role via environment
variables:

```console
export AWS_ROLE_{role alias}=arn:aws:iam:.....:role/SlashInfraInspection
# The region defaults to `us-east-1` if left unspecified
export AWS_REGION_{role alias}=eu-west-2

# e.g.
export AWS_ROLE_PRODUCTION=arn:aws:iam:.....:role/SlashInfraInspection
export AWS_REGION_PRODUCTION=eu-west-2
```

If you need to search multiple regions within a single account you can
create several aliases that use the same role ARN.