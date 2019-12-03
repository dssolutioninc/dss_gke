resource "google_pubsub_topic" "primary" {
  name = var.topic_name
}

resource "google_pubsub_subscription" "primary" {
  name  = var.subscription_name
  topic = google_pubsub_topic.primary.name

  labels = {
    app = var.app
    env = var.env
  }

  message_retention_duration  = var.message_retention_duration
  retain_acked_messages       = var.retain_acked_messages
  ack_deadline_seconds        = var.ack_deadline_seconds
}