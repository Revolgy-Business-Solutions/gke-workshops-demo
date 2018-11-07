resource "google_container_node_pool" "prod-pool" {
  name       = "master-pool"
  zone       = "europe-west1-b"
  cluster    = "${google_container_cluster.primary.name}"
#  initial_node_count = 4
  initial_node_count = 1
  
 # autoscaling {
 #   min_node_count = 4
 #   max_node_count = 8
 # } 

  management {
    auto_repair = "true"
    auto_upgrade = "false"
  } 

  node_config {
  
    machine_type = "n1-highmem-2"

    oauth_scopes = [
      "https://www.googleapis.com/auth/compute",
      "storage-rw",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]

    labels {
      foo = "bar"
    }

    tags = ["foo", "bar"]
    taint {
      key = "env"
      value = "prod"
      effect = "NO_SCHEDULE"
    }



resource "google_container_node_pool" "dev-pool" {
  name       = "preemptible-pool"
  zone       = "europe-west1-b"
  
  cluster    = "${google_container_cluster.primary.name}"
#  initial_node_count = 2
  initial_node_count = 1

#  autoscaling {
#    min_node_count = 2
#    max_node_count = 6
#  }

  management {
    auto_repair = "true"
    auto_upgrade = "false"
  }
  
  node_config {
    oauth_scopes = [
      "compute-rw",
      "storage-rw",
      "logging-write",
      "monitoring-write",
    ]

    preemptible = true
    machine_type = "n1-highmem-2"

    labels {
      foo = "bar"
    }

    tags = ["foo", "bar"]
  }
}



resource "google_container_cluster" "primary" {
  name              = "prod"
  zone              = "europe-west1-b"
  remove_default_node_pool = true

  min_master_version = "1.10.2" 
  #monitoring_service = "monitoring.googleapis.com/kubernetes"
  #logging_service    = "logging.googleapis.com/kubernetes" 
  additional_zones = [
    "europe-west1-c",
  ]

  lifecycle {
    ignore_changes = ["node_pool"]
  }

  node_pool {
    name = "default-pool"
  }
}



