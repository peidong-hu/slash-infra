resource "aws_iam_role" "slash-infra-access" {
  name               = "${var.role_name}"
  assume_role_policy = "${data.aws_iam_policy_document.allow-slash-infra-account-to-assume.json}"

  tags = {
    CreatedBy = "terraform"
  }
}

data "aws_iam_policy_document" "allow-slash-infra-account-to-assume" {
  statement {
    id      = "1"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = ["${var.trusted_aws_account_id}"]
    }
  }
}
