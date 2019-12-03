terraform {
  required_version = "~>0.12.14"
}

## project ##
provider "google" {
  project     = var.project
  region      = var.region
}

## storage buckets ##
module "storage" {
  source = "./_modules/storage"

  name   = var.storage_name
  region = var.region
  app    = var.app
  env    = var.env
}

## cluster ##
module "cluster" {
  source = "./_modules/cluster"

  name  = var.cluster_name
  zone  = var.zone
  node_count = var.node_count
  initial_node_count = var.initial_node_count
  node_pool_name = var.node_pool_name
  node_machine_type = var.node_machine_type
}

## topic & subscription ##
## the first one
module "the-first-topic" {
  source = "./_modules/pubsub"

  topic_name  = var.first_topic_name
  subscription_name  = var.first_subscription_name

  app    = var.app
  env    = var.env
}

## the second one
module "the-second-topic" {
  source = "./_modules/pubsub"

  topic_name  = var.second_topic_name
  subscription_name  = var.second_subscription_name

  app    = var.app
  env    = var.env
}
