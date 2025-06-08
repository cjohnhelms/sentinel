terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket = "sentinel-prod-s3bucket-use1"
    key    = "tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = "us-east-1"
  profile = "terraform"
  default_tags {
    Application = "sentinel"
    Envrionment = "production"
  }
}

// backend config
import {
  to = aws_s3_bucket.backend
  id = "sentinel-prod-s3bucket-use1"
}

resource "aws_s3_bucket" "backend" {
  bucket = "sentinel-prod-s3bucket-use1"
}

resource "aws_s3_bucket_versioning" "backend" {
  bucket = "sentinel-prod-s3bucket-use1"
  versioning_configuration {
    status = "Enabled"
  }
}

// lambda role
resource "aws_iam_role" "lambda_role" {
  name               = "sentinel-prod-lambda-role-use1"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

// lambda
resource "aws_lambda_function" "lambda" {
  filename      = "../lambda_function_${var.lambdasVersion}.zip"
  function_name = "sentinel-prod-lambda-use1"
  role          = aws_iam_role.lambda_role.arn
  handler       = "index.handler"
  runtime       = "nodejs18.x"
  memory_size   = 1024
  timeout       = 300
}

// eventbridge
resource "aws_scheduler_schedule_group" "schedule-group" {
  name = "sentinel-prod-eventbridge-schedule-group-use1"
}

resource "aws_scheduler_schedule" "schedule" {
  name = "sentinel-prod-eventbridge-schedule-use1"
  group_name = "sentinel-prod-eventbridge-schedule-group-use1"

  schedule_expression = "cron(0 0 * * * *)"
  schedule_expression_timezone = "America/Chicago"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_lambda_function.lambda.arn
    role_arn = aws_iam_role.lambda_role.arn
  }
}