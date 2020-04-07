resource "aws_dynamodb_table" "link" {
  name           = "Link"
  billing_mode   = "PROVISIONED"
  hash_key       = "ShortLink"
  read_capacity  = 20
  write_capacity = 20

  attribute {
    name = "ShortLink"
    type = "S"
  }

  attribute {
    name = "Hash"
    type = "S"
  }




  global_secondary_index {
    name            = "HashIndex"
    hash_key        = "Hash"
    projection_type = "KEYS_ONLY"
    read_capacity   = 20
    write_capacity  = 20
  }

}

# POLICIES
resource "aws_iam_role_policy" "dynamodb-lambda-policy" {
  name   = "dynamodb_lambda_policy"
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
      "Resource": "${aws_dynamodb_table.link.arn}"
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
      "Resource": "${aws_dynamodb_table.link.arn}/index/*"
    }
  ]
}
EOF
}

