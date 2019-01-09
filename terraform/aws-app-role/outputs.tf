output "iam_role_arn" {
  description = "The ARN of the IAM role that will be assumed by slash-infra's IAM user. Useful for attaching additional policies to the role that are specific to your org"
  value       = "${aws_iam_role.slash-infra-access.arn}"
}
