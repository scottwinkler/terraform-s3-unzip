provider "aws" {
    profile = "em-publiccloud-dev"
    region = "us-west-2"
}

module "test" {
    source = ".."
    src_bucket = "${aws_s3_bucket.s3_bucket.bucket}"
    src_bucket_arn = "${aws_s3_bucket.s3_bucket.arn}"
    project_name = "test"
}

resource "aws_s3_bucket" "s3_bucket" {
  bucket        = "s3-unzip-test-098"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "bucket_object" {
  bucket = "${aws_s3_bucket.s3_bucket.bucket}"
  key    = "t/test.zip"
  source = "${path.cwd}/test.zip"
   tags = {
    unzip_id = "${module.test.id}"  
  }
}