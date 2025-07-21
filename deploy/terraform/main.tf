terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket = "sentinel-prod-s3-bucket-use1"
    key    = "tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "terraform"
  default_tags {
    tags = {
      Project     = "sentinel"
      Environment = "production"
      Region      = "us-east-1"
    }
  }
}

// backend config
import {
  to = aws_s3_bucket.backend
  id = "sentinel-prod-s3-bucket-use1"
}

resource "aws_s3_bucket" "backend" {
  bucket = "sentinel-prod-s3-bucket-use1"
}

resource "aws_s3_bucket_versioning" "backend" {
  bucket = "sentinel-prod-s3-bucket-use1"
  versioning_configuration {
    status = "Enabled"
  }
}

// lambda assume policy
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "sentinel-prod-iam-role-lambda-use1"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

// eventbridge assume policy
data "aws_iam_policy_document" "eventbridge_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      identifiers = ["scheduler.amazonaws.com"]
      type        = "Service"
    }
  }
}

// invoke permissions
data "aws_iam_policy_document" "eventbridge" {
  statement {
    actions   = ["lambda:InvokeFunction"]
    effect    = "Allow"
    resources = [aws_lambda_function.lambda.arn]
  }
}

resource "aws_iam_policy" "eventbridge" {
  name   = "sentinel-prod-iam-policy-eventbridge-use1"
  policy = data.aws_iam_policy_document.eventbridge.json
}

resource "aws_iam_role_policy_attachment" "eventbridge" {
  role       = aws_iam_role.eventbridge.name
  policy_arn = aws_iam_policy.eventbridge.arn
}

resource "aws_iam_role" "eventbridge" {
  name               = "sentinel-prod-iam-role-eventbridge-use1"
  assume_role_policy = data.aws_iam_policy_document.eventbridge_assume.json
}

// lambda
resource "aws_lambda_function" "lambda" {
  filename      = "./lambda_${var.githubSHA}.zip"
  function_name = "sentinel-prod-lambda-function-use1"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  memory_size   = 128
  timeout       = 3

  source_code_hash = filebase64sha256("lambda_${var.githubSHA}.zip")
}

// eventbridge
resource "aws_scheduler_schedule_group" "schedule_group" {
  name = "sentinel-prod-eventbridge-group-use1"
}

resource "aws_scheduler_schedule" "schedule" {
  name       = "sentinel-prod-eventbridge-schedule-use1"
  group_name = "sentinel-prod-eventbridge-group-use1"

  schedule_expression          = "cron(0 9 * * ? *)"
  schedule_expression_timezone = "America/Chicago"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_lambda_function.lambda.arn
    role_arn = aws_iam_role.eventbridge.arn
  }
}

// logging
resource "aws_cloudwatch_log_group" "log_group" {
  name = "/aws/lambda/sentinel-prod-lambda-function-use1"

  retention_in_days = 1
}

data "aws_iam_policy" "lambda_basic" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_logging_policy_role_attachment" {
  role       = aws_iam_role.lambda.name
  policy_arn = data.aws_iam_policy.lambda_basic.arn
}