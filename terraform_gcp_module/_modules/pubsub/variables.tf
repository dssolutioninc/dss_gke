variable "topic_name" {}
variable "subscription_name" {}
variable "app" {}
variable "env" {}


variable "message_retention_duration" {
  default = "604800s" # 7 days
}

variable "retain_acked_messages" {
  default = true
}

variable "ack_deadline_seconds" {
  default = 20
}