variable "grafana_auth" {
  description = "The authentication token for the Grafana instance"
  type        = string

}

variable "customer_slug" {
  description = "The slug for the customer"
  type        = string

}

variable "customer_name" {
  description = "The name of the customer"
  type        = string
}

variable "env" {
  description = "The environment"
  type        = string

}

variable "stack_region" {
  description = "The region for the stack"
  type        = string

}

variable "cloud_sa_token_name" {
  description = "The name of the service account"
  type        = string
  default     = "customer_stack cloud_sa key"

}
