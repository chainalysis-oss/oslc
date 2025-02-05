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

