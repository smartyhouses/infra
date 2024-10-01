// Step 1: Create a stack
provider "grafana" {
  alias                     = "cloud"
  cloud_access_policy_token = var.grafana_auth
}

resource "grafana_cloud_stack" "customer_stack" {
  provider = grafana.cloud

  name        = var.customer_name + var.env
  slug        = var.customer_slug + var.env
  region_slug = var.stack_region
}

// Step 2: Create a service account and key for the stack
resource "grafana_cloud_stack_service_account" "cloud_sa" {
  provider   = grafana.cloud
  stack_slug = grafana_cloud_stack.customer_stack.slug

  name        = "cloud service account"
  role        = "Admin"
  is_disabled = false
}

resource "grafana_cloud_stack_service_account_token" "cloud_sa" {
  provider   = grafana.cloud
  stack_slug = grafana_cloud_stack.customer_stack.slug

  name               = var.cloud_sa_token_name
  service_account_id = grafana_cloud_stack_service_account.cloud_sa.id
}

// Step 3: Create resources within the stack
provider "grafana" {
  alias = var.customer_name + var.env

  url  = grafana_cloud_stack.customer_stack.url
  auth = grafana_cloud_stack_service_account_token.cloud_sa.key
}

resource "grafana_folder" "my_folder" {
  provider = grafana.customer_stack

  title = "Test Folder"
}
