variable "src_bucket" {
  type = "string"
  description = "required - The source bucket to listen for put object events"
}

variable "src_bucket_arn" {
  type = "string"
  description = "required - The arn of the same source bucket"
}

variable "dst_bucket" {
  type = "string"
  default = ""
  description = "optional - The destination bucket to send the unzipped files, if not the source bucket"
}

variable "project_name" {
  type    = "string"
  default = ""
  description = "optional - identifier for your project"
}

variable "delete_zip" {
  type = "string"
  default = false
  description = "optional - set true if you wish to delete the original zip file that is unzipped"
}