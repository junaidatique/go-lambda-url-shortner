resource "aws_dynamodb_table" "linkAnalytics" {
  name           = "LinkAnalytics"
  billing_mode   = "PROVISIONED"
  hash_key       = "RequestID"
  read_capacity  = 20
  write_capacity = 20

  attribute {
    name = "RequestID"
    type = "S"
  }

  attribute {
    name = "ShortLink"
    type = "S"
  }

  global_secondary_index {
    name            = "ShortLinkIndex"
    hash_key        = "ShortLink"
    projection_type = "ALL"
    read_capacity   = 20
    write_capacity  = 20
  }

}

# POLICIES
resource "aws_iam_role_policy" "dynamodb-lambda-policy-linkAnalytics" {
  name   = "dynamodb_lambda_policy_linkAnalytics"
  role   = aws_iam_role.lambda_exec.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:Query",
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": "${aws_dynamodb_table.linkAnalytics.arn}"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:Query",
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": "${aws_dynamodb_table.linkAnalytics.arn}/index/*"
    }
  ]
}
EOF
}

