//if you need to put a dependency on this module, use this output
output "id" {
    depends_on = [
        "aws_s3_bucket_notification.bucket_notification",
        "aws_lambda_permission.allow_bucket"
    ]
    value = "${random_pet.random.id}"
}