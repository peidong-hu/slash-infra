resource "aws_iam_user" "app-user" {
  name = var.username
}

resource "aws_iam_user_policy" "allow-assuming-slash-infra-roles" {
  name = "allow-assuming-slash-infra-roles"
  user = aws_iam_user.app-user.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Sid": "AllowAssumingSlashInfraRoles",
          "Effect": "Allow",
          "Action": "sts:AssumeRole",
          "Resource": "arn:aws:iam::*:role/${var.role_name}"
      }
  ]
}
EOF
}
