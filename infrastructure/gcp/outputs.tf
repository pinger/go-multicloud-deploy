output "trigger_url" {
  value       = google_cloudfunctions_function.function.https_trigger_url
  description = "URL which triggers function execution."
}