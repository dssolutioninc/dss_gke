resource "google_storage_bucket" "sample-bucket" {
  name          = var.name
  location      = var.region
  storage_class = var.storage_class

  labels = {
    app = var.app
    env = var.env
  }
}