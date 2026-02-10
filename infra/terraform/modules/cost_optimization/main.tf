# VirtEngine Cost Optimization Module
#
# This module implements comprehensive infrastructure cost optimization including:
# - Cost allocation tags and tagging policies
# - AWS Budgets with alerts
# - Cost anomaly detection
# - Unused resource cleanup automation
# - Savings plan recommendations
# - Cost optimization recommendations

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Local Values
# -----------------------------------------------------------------------------

locals {
  common_tags = merge(var.tags, {
    Project     = "virtengine"
    Environment = var.environment
    ManagedBy   = "terraform"
    CostCenter  = var.cost_center
    Owner       = var.owner
    Module      = "cost-optimization"
  })

  # Budget alert thresholds
  budget_thresholds = [50, 80, 100, 120]

  # Resource cleanup retention periods (days)
  cleanup_retention = {
    snapshots          = 30
    unattached_volumes = 7
    old_amis           = 90
    unused_eips        = 3
  }
}

# -----------------------------------------------------------------------------
# Cost Allocation Tags
# -----------------------------------------------------------------------------

# Enable cost allocation tags in the account
resource "aws_ce_cost_allocation_tag" "project" {
  tag_key = "Project"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "environment" {
  tag_key = "Environment"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "cost_center" {
  tag_key = "CostCenter"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "owner" {
  tag_key = "Owner"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "service" {
  tag_key = "Service"
  status  = "Active"
}

resource "aws_ce_cost_allocation_tag" "team" {
  tag_key = "Team"
  status  = "Active"
}

# -----------------------------------------------------------------------------
# AWS Budgets
# -----------------------------------------------------------------------------

# Monthly total budget
resource "aws_budgets_budget" "monthly_total" {
  name         = "virtengine-${var.environment}-monthly-total"
  budget_type  = "COST"
  limit_amount = var.monthly_budget_limit
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "TagKeyValue"
    values = ["Project$virtengine", "Environment$${var.environment}"]
  }

  dynamic "notification" {
    for_each = local.budget_thresholds
    content {
      comparison_operator        = notification.value <= 100 ? "GREATER_THAN" : "GREATER_THAN"
      threshold                  = notification.value
      threshold_type             = notification.value <= 100 ? "PERCENTAGE" : "PERCENTAGE"
      notification_type          = notification.value <= 100 ? "ACTUAL" : "FORECASTED"
      subscriber_email_addresses = var.budget_notification_emails
      subscriber_sns_topic_arns  = var.budget_notification_sns_arns
    }
  }
}

# Compute budget
resource "aws_budgets_budget" "compute" {
  name         = "virtengine-${var.environment}-compute"
  budget_type  = "COST"
  limit_amount = var.compute_budget_limit
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Elastic Compute Cloud - Compute", "Amazon Elastic Container Service for Kubernetes"]
  }

  cost_filter {
    name   = "TagKeyValue"
    values = ["Project$virtengine", "Environment$${var.environment}"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = var.budget_notification_emails
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = var.budget_notification_emails
  }
}

# Storage budget
resource "aws_budgets_budget" "storage" {
  name         = "virtengine-${var.environment}-storage"
  budget_type  = "COST"
  limit_amount = var.storage_budget_limit
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Simple Storage Service", "Amazon Elastic Block Store"]
  }

  cost_filter {
    name   = "TagKeyValue"
    values = ["Project$virtengine", "Environment$${var.environment}"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = var.budget_notification_emails
  }
}

# Networking budget
resource "aws_budgets_budget" "networking" {
  name         = "virtengine-${var.environment}-networking"
  budget_type  = "COST"
  limit_amount = var.networking_budget_limit
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Virtual Private Cloud", "AWS Global Accelerator", "Amazon Route 53"]
  }

  cost_filter {
    name   = "TagKeyValue"
    values = ["Project$virtengine", "Environment$${var.environment}"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = var.budget_notification_emails
  }
}

# Data transfer budget
resource "aws_budgets_budget" "data_transfer" {
  name         = "virtengine-${var.environment}-data-transfer"
  budget_type  = "COST"
  limit_amount = var.data_transfer_budget_limit
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "UsageType"
    values = ["DataTransfer-Out-Bytes", "DataTransfer-Regional-Bytes"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = var.budget_notification_emails
  }
}

# -----------------------------------------------------------------------------
# Cost Anomaly Detection
# -----------------------------------------------------------------------------

# Cost anomaly monitor for the project
resource "aws_ce_anomaly_monitor" "project" {
  name              = "virtengine-${var.environment}-anomaly-monitor"
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
}

# Cost anomaly monitor for linked accounts (if using AWS Organizations)
resource "aws_ce_anomaly_monitor" "linked_accounts" {
  count = var.enable_organization_monitor ? 1 : 0

  name              = "virtengine-organization-anomaly-monitor"
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "LINKED_ACCOUNT"
}

# Cost anomaly subscription for alerts
resource "aws_ce_anomaly_subscription" "main" {
  name = "virtengine-${var.environment}-anomaly-subscription"

  frequency = var.anomaly_detection_frequency

  monitor_arn_list = compact([
    aws_ce_anomaly_monitor.project.arn,
    var.enable_organization_monitor ? aws_ce_anomaly_monitor.linked_accounts[0].arn : ""
  ])

  subscriber {
    type    = "EMAIL"
    address = var.anomaly_notification_email
  }

  dynamic "subscriber" {
    for_each = var.anomaly_notification_sns_arn != "" ? [1] : []
    content {
      type    = "SNS"
      address = var.anomaly_notification_sns_arn
    }
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      match_options = ["GREATER_THAN_OR_EQUAL"]
      values        = [tostring(var.anomaly_threshold_absolute)]
    }
  }
}

# -----------------------------------------------------------------------------
# SNS Topic for Cost Alerts
# -----------------------------------------------------------------------------

resource "aws_sns_topic" "cost_alerts" {
  name = "virtengine-${var.environment}-cost-alerts"

  tags = merge(local.common_tags, {
    Name = "virtengine-${var.environment}-cost-alerts"
  })
}

resource "aws_sns_topic_policy" "cost_alerts" {
  arn = aws_sns_topic.cost_alerts.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowBudgetNotifications"
        Effect = "Allow"
        Principal = {
          Service = "budgets.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.cost_alerts.arn
      },
      {
        Sid    = "AllowCostAnomalyDetection"
        Effect = "Allow"
        Principal = {
          Service = "costalerts.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.cost_alerts.arn
      }
    ]
  })
}

resource "aws_sns_topic_subscription" "cost_alert_email" {
  for_each = toset(var.budget_notification_emails)

  topic_arn = aws_sns_topic.cost_alerts.arn
  protocol  = "email"
  endpoint  = each.value
}

# -----------------------------------------------------------------------------
# Lambda Function for Unused Resource Cleanup
# -----------------------------------------------------------------------------

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# IAM Role for cleanup Lambda
resource "aws_iam_role" "cleanup_lambda" {
  count = var.enable_resource_cleanup ? 1 : 0

  name = "virtengine-${var.environment}-cleanup-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy" "cleanup_lambda" {
  count = var.enable_resource_cleanup ? 1 : 0

  name = "cleanup-policy"
  role = aws_iam_role.cleanup_lambda[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "EC2ReadAccess"
        Effect = "Allow"
        Action = [
          "ec2:DescribeVolumes",
          "ec2:DescribeSnapshots",
          "ec2:DescribeImages",
          "ec2:DescribeAddresses",
          "ec2:DescribeInstances",
          "ec2:DescribeTags"
        ]
        Resource = "*"
      },
      {
        Sid    = "EC2CleanupAccess"
        Effect = "Allow"
        Action = [
          "ec2:DeleteVolume",
          "ec2:DeleteSnapshot",
          "ec2:DeregisterImage",
          "ec2:ReleaseAddress"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:ResourceTag/Environment" = var.environment
            "aws:ResourceTag/Project"     = "virtengine"
          }
        }
      },
      {
        Sid      = "SNSPublish"
        Effect   = "Allow"
        Action   = ["sns:Publish"]
        Resource = [aws_sns_topic.cost_alerts.arn]
      },
      {
        Sid    = "CloudWatchLogs"
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Sid    = "CostExplorerAccess"
        Effect = "Allow"
        Action = [
          "ce:GetCostAndUsage",
          "ce:GetReservationUtilization",
          "ce:GetSavingsPlansCoverage",
          "ce:GetSavingsPlansUtilization",
          "ce:GetRightsizingRecommendation"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cleanup_lambda_basic" {
  count = var.enable_resource_cleanup ? 1 : 0

  role       = aws_iam_role.cleanup_lambda[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Lambda function for resource cleanup
resource "aws_lambda_function" "resource_cleanup" {
  count = var.enable_resource_cleanup ? 1 : 0

  function_name = "virtengine-${var.environment}-resource-cleanup"
  role          = aws_iam_role.cleanup_lambda[0].arn
  handler       = "index.handler"
  runtime       = "python3.11"
  timeout       = 300
  memory_size   = 256

  filename         = data.archive_file.cleanup_lambda[0].output_path
  source_code_hash = data.archive_file.cleanup_lambda[0].output_base64sha256

  environment {
    variables = {
      ENVIRONMENT             = var.environment
      SNS_TOPIC_ARN           = aws_sns_topic.cost_alerts.arn
      DRY_RUN                 = var.cleanup_dry_run ? "true" : "false"
      SNAPSHOT_RETENTION_DAYS = local.cleanup_retention.snapshots
      VOLUME_RETENTION_DAYS   = local.cleanup_retention.unattached_volumes
      AMI_RETENTION_DAYS      = local.cleanup_retention.old_amis
      EIP_RETENTION_DAYS      = local.cleanup_retention.unused_eips
    }
  }

  tags = local.common_tags
}

# Lambda source code
data "archive_file" "cleanup_lambda" {
  count = var.enable_resource_cleanup ? 1 : 0

  type        = "zip"
  output_path = "${path.module}/files/cleanup_lambda.zip"

  source {
    content  = file("${path.module}/files/cleanup_lambda.py")
    filename = "index.py"
  }
}

# CloudWatch Event Rule to trigger cleanup weekly
resource "aws_cloudwatch_event_rule" "cleanup_schedule" {
  count = var.enable_resource_cleanup ? 1 : 0

  name                = "virtengine-${var.environment}-resource-cleanup"
  description         = "Trigger resource cleanup Lambda weekly"
  schedule_expression = var.cleanup_schedule

  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "cleanup_lambda" {
  count = var.enable_resource_cleanup ? 1 : 0

  rule      = aws_cloudwatch_event_rule.cleanup_schedule[0].name
  target_id = "CleanupLambda"
  arn       = aws_lambda_function.resource_cleanup[0].arn
}

resource "aws_lambda_permission" "cleanup_cloudwatch" {
  count = var.enable_resource_cleanup ? 1 : 0

  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.resource_cleanup[0].function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.cleanup_schedule[0].arn
}

# -----------------------------------------------------------------------------
# Lambda Function for Cost Recommendations
# -----------------------------------------------------------------------------

resource "aws_lambda_function" "cost_recommendations" {
  count = var.enable_cost_recommendations ? 1 : 0

  function_name = "virtengine-${var.environment}-cost-recommendations"
  role          = aws_iam_role.cleanup_lambda[0].arn
  handler       = "index.handler"
  runtime       = "python3.11"
  timeout       = 300
  memory_size   = 256

  filename         = data.archive_file.recommendations_lambda[0].output_path
  source_code_hash = data.archive_file.recommendations_lambda[0].output_base64sha256

  environment {
    variables = {
      ENVIRONMENT   = var.environment
      SNS_TOPIC_ARN = aws_sns_topic.cost_alerts.arn
    }
  }

  tags = local.common_tags
}

data "archive_file" "recommendations_lambda" {
  count = var.enable_cost_recommendations ? 1 : 0

  type        = "zip"
  output_path = "${path.module}/files/recommendations_lambda.zip"

  source {
    content  = file("${path.module}/files/cost_recommendations_lambda.py")
    filename = "index.py"
  }
}

# CloudWatch Event Rule to generate recommendations monthly
resource "aws_cloudwatch_event_rule" "recommendations_schedule" {
  count = var.enable_cost_recommendations ? 1 : 0

  name                = "virtengine-${var.environment}-cost-recommendations"
  description         = "Generate cost recommendations monthly"
  schedule_expression = "cron(0 9 1 * ? *)" # 1st of every month at 9 AM

  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "recommendations_lambda" {
  count = var.enable_cost_recommendations ? 1 : 0

  rule      = aws_cloudwatch_event_rule.recommendations_schedule[0].name
  target_id = "RecommendationsLambda"
  arn       = aws_lambda_function.cost_recommendations[0].arn
}

resource "aws_lambda_permission" "recommendations_cloudwatch" {
  count = var.enable_cost_recommendations ? 1 : 0

  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cost_recommendations[0].function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.recommendations_schedule[0].arn
}

# -----------------------------------------------------------------------------
# CloudWatch Dashboard for Cost Monitoring
# -----------------------------------------------------------------------------

resource "aws_cloudwatch_dashboard" "cost" {
  dashboard_name = "virtengine-${var.environment}-cost-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "text"
        x      = 0
        y      = 0
        width  = 24
        height = 1
        properties = {
          markdown = "# VirtEngine Cost Dashboard - ${upper(var.environment)}"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 1
        width  = 8
        height = 6
        properties = {
          title  = "EC2 Running Instances"
          region = data.aws_region.current.name
          stat   = "Average"
          period = 3600
          metrics = [
            ["AWS/EC2", "CPUUtilization", { stat = "SampleCount", label = "Running Instances" }]
          ]
        }
      },
      {
        type   = "metric"
        x      = 8
        y      = 1
        width  = 8
        height = 6
        properties = {
          title  = "NAT Gateway Data Transfer"
          region = data.aws_region.current.name
          stat   = "Sum"
          period = 3600
          metrics = [
            ["AWS/NATGateway", "BytesOutToDestination"],
            [".", "BytesOutToSource"],
            [".", "BytesInFromDestination"],
            [".", "BytesInFromSource"]
          ]
        }
      },
      {
        type   = "metric"
        x      = 16
        y      = 1
        width  = 8
        height = 6
        properties = {
          title  = "S3 Bucket Size"
          region = data.aws_region.current.name
          stat   = "Average"
          period = 86400
          metrics = [
            ["AWS/S3", "BucketSizeBytes", "StorageType", "StandardStorage"]
          ]
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 7
        width  = 12
        height = 6
        properties = {
          title  = "EKS Node CPU Utilization"
          region = data.aws_region.current.name
          stat   = "Average"
          period = 300
          metrics = [
            ["ContainerInsights", "node_cpu_utilization", "ClusterName", var.cluster_name]
          ]
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 7
        width  = 12
        height = 6
        properties = {
          title  = "EKS Node Memory Utilization"
          region = data.aws_region.current.name
          stat   = "Average"
          period = 300
          metrics = [
            ["ContainerInsights", "node_memory_utilization", "ClusterName", var.cluster_name]
          ]
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 13
        width  = 24
        height = 6
        properties = {
          title  = "EBS Volume IOPS"
          region = data.aws_region.current.name
          stat   = "Average"
          period = 300
          metrics = [
            ["AWS/EBS", "VolumeReadOps"],
            [".", "VolumeWriteOps"]
          ]
        }
      }
    ]
  })
}
