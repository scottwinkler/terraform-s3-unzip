resource "random_pet" "random" {}

locals {
  module_name = "s3_unzip"
  dst_bucket = "${var.dst_bucket == "" ? var.src_bucket : var.dst_bucket}"
}

#lambda role
data "aws_iam_policy_document" "lambda_assume_role_policy_document" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lambda_iam_policy_document" {
  statement {
      effect = "Allow"
      actions = ["s3:GetObject","s3:DeleteObject","s3:PutObject","logs:*"]
      resources = ["*"] 
  }
}

resource "aws_iam_policy" "lambda_iam_policy" {
  name = "${var.project_name}-${local.module_name}-${random_pet.random.id}"
  path = "/"
  policy = "${data.aws_iam_policy_document.lambda_iam_policy_document.json}"
}

resource "aws_iam_role" "lambda_role" {
  name               = "${var.project_name}-${local.module_name}-${random_pet.random.id}"
  path               = "/"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role_policy_document.json}"
}

resource "aws_iam_policy_attachment" "lambda_iam_policy_role_attachment" {
  name = "${var.project_name}-${local.module_name}-${random_pet.random.id}"
  roles = ["${aws_iam_role.lambda_role.name}"]
  policy_arn = "${aws_iam_policy.lambda_iam_policy.arn}"
}

resource "aws_lambda_function" "lambda_function" {
  filename = "${path.module}/golang/deployment.zip"
  function_name = "${var.project_name}-${local.module_name}-${random_pet.random.id}"
  description = "A helper lambda function to unzip files from an s3 bucket"
  handler       = "deployment"
  role          = "${aws_iam_role.lambda_role.arn}"
  memory_size   = 256
  runtime       = "go1.x"
  timeout       = 60
  environment {
    variables = {
      DST_BUCKET = "${local.dst_bucket}"
      DELETE_ZIP = "${var.delete_zip}"
    }
  }
}	

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = "${var.src_bucket}"
  lambda_function {
    lambda_function_arn = "${aws_lambda_function.lambda_function.arn}"
    events              = ["s3:ObjectCreated:*"]
  }
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.lambda_function.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${var.src_bucket_arn}"
}