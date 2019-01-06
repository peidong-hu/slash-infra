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

This approach seems convoluted, but there are several benefits.

Firstly, when an IAM user "assumes" a role, AWS generates a set of temporary
credentials for the user that have the same permissions as the role.
These credentials are short-lived, and are rotated transparently by
the AWS SDK. If these temporary credentials were leaked to a
third party, they would only be usable for a short period of time.

Secondly, if the credentials for the IAM user are leaked, the attacker
will only be able to assume IAM roles. If they do not know the ARN of
your role they will not be able to assume it, and thus won't be able to
perform actions on your account.

Note that if you're using the role name suggested in these docs then
they will likely be able to guess the full ARN, as you can always get
the ID of the AWS account credentials belong to using `aws sts
get-caller-identity`. If this is a concern for you, you can make things
slightly more difficult for attackers by choosing a unique name for your
roles that are different to the name of the IAM user. `slash-infra` only
uses the role ARN to authenticate to AWS, so you could use a different
role name for each environment.

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

## Testing locally

Download [ngrok](http://ngrok.com), and [create a slack
app](https://api.slack.com/apps) in your slack workspace. Create slash
commands in the app for the commands you want to support (see server.go).

## FAQ

### Why not write this in a lambda?

- I don't know how to write lambdas. I wrote this in 20% time and didn't
  want to spend my day learning how to deploy lambdas
- We don't mind spending the few $X heroku charge to run this on a hobby
  dyno
