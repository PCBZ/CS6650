# Lambda Function for Order Processing
# Subscribes directly to SNS topic (no SQS)

# Use existing LabRole (AWS Academy doesn't allow creating IAM roles)
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Lambda function
resource "aws_lambda_function" "order_processor" {
  count         = var.enable_lambda ? 1 : 0
  filename      = "${path.root}/../src/lambda/bootstrap.zip"
  function_name = "${var.service_name}-order-processor-lambda"
  role          = data.aws_iam_role.lab_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"  # Go runtime
  memory_size   = 512              # 512MB as required
  timeout       = 30               # 30 seconds (needs time for 3-second processing)

  source_code_hash = filebase64sha256("${path.root}/../src/lambda/bootstrap.zip")

  # No custom environment variables needed - AWS_REGION is automatically set by Lambda

  tags = {
    Name = "${var.service_name}-order-processor-lambda"
  }
}

# SNS subscription for Lambda (direct trigger, no SQS)
resource "aws_sns_topic_subscription" "lambda" {
  count     = var.enable_lambda ? 1 : 0
  topic_arn = var.sns_topic_arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.order_processor[0].arn
}

# Permission for SNS to invoke Lambda
resource "aws_lambda_permission" "sns" {
  count         = var.enable_lambda ? 1 : 0
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.order_processor[0].function_name
  principal     = "sns.amazonaws.com"
  source_arn    = var.sns_topic_arn
}

# CloudWatch Log Group for Lambda
resource "aws_cloudwatch_log_group" "lambda" {
  count             = var.enable_lambda ? 1 : 0
  name              = "/aws/lambda/${var.service_name}-order-processor-lambda"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.service_name}-lambda-logs"
  }
}

# Outputs
output "lambda_function_name" {
  value = var.enable_lambda ? aws_lambda_function.order_processor[0].function_name : null
}

output "lambda_function_arn" {
  value = var.enable_lambda ? aws_lambda_function.order_processor[0].arn : null
}
