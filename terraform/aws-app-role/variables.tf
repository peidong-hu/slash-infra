variable "role_name" {
  description = "The name of the role to create. This should match the name of the role your IAM user has permission to assume"
  default     = "SlashInfraAccess"
}

variable "trusted_aws_account_id" {
  description = "The ID of the account in which the AWS IAM user for slash-infra lives"
}
