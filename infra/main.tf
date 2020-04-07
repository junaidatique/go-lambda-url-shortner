provider "aws" {
  region = "us-east-1"  
}


variable "app_name" {
  description = "App Name"
  default = "lambda-url-shortner"
}

variable "app_env" {
  description = "Application Env Tag"
  default = "dev"
}

locals {
  app_id = "${lower(var.app_name)}-${lower(var.app_env)}-${random_id.unique_suffix.hex}"
  link_app_name = "${lower(var.app_name)}"
}

data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "build/bin/app"
  output_path = "build/bin/app.zip"
}

resource "random_id" "unique_suffix" {
  byte_length = 2
}

output "app_url" {
  value = aws_api_gateway_deployment.api_deployment.invoke_url
}



