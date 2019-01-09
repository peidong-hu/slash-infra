variable "username" {
  description = "This will be the username given to the AWS IAM user"
  default     = "slash-infra-app"
}

variable "role_name" {
  description = "The name of the role your AWS accounts will use. By default this is the only role the user will be allowed to assume"
  default     = "SlashInfraApp"
}
