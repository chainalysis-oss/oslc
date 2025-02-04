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

