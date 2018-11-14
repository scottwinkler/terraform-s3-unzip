#terraform-s3-unzip
This is a terraform module for unzipping a file from put to an s3 bucket. It creates a lambda function, role, policy and s3 bucket notification trigger. When a zip file is uploaded to the src s3 bucket, then it will unzip it, and upload those files to the dst bucket.

## Argument Reference

* src_bucket - (Required) The source bucket to listen for put object events
* src_bucket_arn - (Required) The arn of the same source bucket
* project_name - (Optional) Identifier for your project
* dst_bucket - (Optional) he destination bucket to send the unzipped files, if not the source bucket


#Example Usage

module "test" {
    source = "github.com/scottwinkler/terraform-s3-unzip"
    src_bucket = "${aws_s3_bucket.s3_bucket.bucket}"
    src_bucket_arn = "${aws_s3_bucket.s3_bucket.arn}"
    project_name = "test"
}

Credit to https://github.com/toshi0607/s3-unzipper-go for writing the go code for a similar idea.