## global variables
variable "project" {}
variable "app" {}
variable "env" {}
variable "region" {
  default = "asia-northeast1"
}
variable "zone" {
  default = "asia-northeast1-b"
}

## storage variables
variable "storage_name" {}

## cluster variables
variable "cluster_name" {}
variable "node_count" {}
variable "initial_node_count" {}
variable "node_pool_name" {}
variable "node_machine_type" {}

## pubsub variables
variable "first_topic_name" {}
variable "first_subscription_name" {}
variable "second_topic_name" {}
variable "second_subscription_name" {}
