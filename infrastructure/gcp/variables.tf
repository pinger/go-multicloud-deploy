variable "region" {
  description = "The Google region to deploy to"
  type        = string
  default     = "us-east1"
}

variable "function_name" {
  description = "The name of the function to provision"
  default     = "bugoga"
}

variable "project" {
  type = string
  default = "project-3289145"
}

variable "function_entry_point" {
  type = string
  default = "EndPoint01"
}