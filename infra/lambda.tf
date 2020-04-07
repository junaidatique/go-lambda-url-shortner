variable "REDIS_URL_HOST" {}
variable "REDIS_PASSWORD" {}
variable "REDIS_DB" {}
variable "URL_COUNTER_KEY" {}
variable "BASE_URL" {}
variable "COUNTER_START_VALUE" {}


resource "aws_lambda_function" "lambda_func" {
  filename         = data.archive_file.lambda_zip.output_path
  function_name    = local.app_id
  handler          = "app"
  source_code_hash = base64sha256(data.archive_file.lambda_zip.output_path)
  runtime          = "go1.x"
  role             = aws_iam_role.lambda_exec.arn
  environment {
    variables = {
      REDIS_URL_HOST      = "${var.REDIS_URL_HOST}"
      REDIS_PASSWORD      = "${var.REDIS_PASSWORD}"
      REDIS_DB            = "${var.REDIS_DB}"
      URL_COUNTER_KEY     = "${var.URL_COUNTER_KEY}"
      BASE_URL            = "${var.BASE_URL}"
      COUNTER_START_VALUE = "${var.COUNTER_START_VALUE}"
    }
  }

}

# Assume role setup
resource "aws_iam_role" "lambda_exec" {
  name_prefix = local.app_id

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

# Attach role to Managed Policy
variable "iam_policy_arn" {
  description = "IAM Policy to be attached to role"
  type        = list(string)

  default = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  ]
}

resource "aws_iam_policy_attachment" "role_attach" {
  name       = "policy-${local.app_id}"
  roles      = [aws_iam_role.lambda_exec.id]
  count      = length(var.iam_policy_arn)
  policy_arn = element(var.iam_policy_arn, count.index)
}
