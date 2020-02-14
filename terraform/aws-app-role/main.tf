resource "aws_iam_role" "slash-infra-access" {
  name               = var.role_name
  assume_role_policy = data.aws_iam_policy_document.allow-slash-infra-account-to-assume.json
}

data "aws_iam_policy_document" "allow-slash-infra-account-to-assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [var.trusted_aws_account_arn]
    }
  }
}

resource "aws_iam_role_policy" "allow-read-only-access" {
  name   = "allow-read-only-access"
  role   = aws_iam_role.slash-infra-access.id
  policy = data.aws_iam_policy_document.allow-read-only-access.json
}

data "aws_iam_policy_document" "allow-read-only-access" {
  statement {
    actions   = ["ec2:DescribeInstances"]
    resources = ["*"]
  }
}
