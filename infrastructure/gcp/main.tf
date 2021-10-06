# ---------------------------------------------------------------------------------------------------------------------
# AWS LAMBDA TERRAFORM EXAMPLE
# See test/terraform_aws_lambda_example_test.go for how to write automated tests for this code.
# ---------------------------------------------------------------------------------------------------------------------
provider "google" {
  region = var.region
  project = var.project
}

terraform {
  # This module is now only being tested with Terraform 0.13.x. However, to make upgrading easier, we are setting
  # 0.12.26 as the minimum version, as that version added support for required_providers with source URLs, making it
  # forwards compatible with 0.13.x code.
  required_version = ">= 0.12.26"
}

provider "archive" {
}

data "archive_file" "zip" {
  type        = "zip"
  source_dir  = "../../src/gcp/"
  output_path = "../../src/${var.function_name}.zip"
}

resource "google_storage_bucket" "bucket" {
  name    = "${var.project}-my-new-bucket"
}

# Add source code zip to bucket
resource "google_storage_bucket_object" "bucket_object" {
  # Append file MD5 to force bucket to be recreated
  name   = "${var.function_name}#${data.archive_file.zip.output_md5}"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.zip.output_path
}

# Enable Cloud Functions API
resource "google_project_service" "cf" {
  service = "cloudfunctions.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Enable Cloud Build API
resource "google_project_service" "cb" {
  service = "cloudbuild.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Enable Service Usage API
resource "google_project_service" "su" {
  service = "serviceusage.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Cloud Resource Manager API
resource "google_project_service" "cr" {
  service = "cloudresourcemanager.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Create Cloud Function
resource "google_cloudfunctions_function" "function" {
  name    = var.function_name
  runtime = "go113" # Switch to a different runtime if needed

  available_memory_mb   = 128
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.bucket_object.name
  trigger_http          = true
  entry_point           = var.function_entry_point
}

# Create IAM entry so all users can invoke the function
resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}