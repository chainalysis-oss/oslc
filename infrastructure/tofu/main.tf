terraform {
  required_version = "= 1.8.7"

  required_providers {
    spacelift = {
      source  = "spacelift-io/spacelift"
      version = "~> 1.19.0"
    }
    ovh = {
      source  = "ovh/ovh"
      version = "1.5.0"
    }
  }
}

provider "spacelift" {}

provider "ovh" {
  endpoint = "ovh-eu"
}

resource "ovh_cloud_project_kube" "oslc_primary" {
  service_name = "ded08a2579ef40d98ed234ccb2061ffe"
  name         = "oslc_primary"
  region       = "DE1"
}

resource "ovh_cloud_project_kube_nodepool" "node_pool" {
  service_name  = "ded08a2579ef40d98ed234ccb2061ffe"
  kube_id       = ovh_cloud_project_kube.oslc_primary.id
  name          = "pool-1"
  flavor_name   = "d2-4"
  desired_nodes = 1
  max_nodes     = 2
  min_nodes     = 1
}

resource "ovh_iam_policy" "spacelift-oslc" {
  name        = "spacelift-oslc-service-account"
  description = "Policy associated with managing OSLC via Spacelift"
  identities = [
    "urn:v1:eu:identity:credential:hm874490-ovh/oauth2-6609ec93ff0ec63f"
  ]
  resources = [
    "urn:v1:eu:resource:*"
  ]
  allow = [
    "publicCloudProject:apiovh:kube/*",
    "publicCloudProject:apiovh:kube/nodepool/*",
    "account:apiovh:iam/policy/*"
  ]
}
