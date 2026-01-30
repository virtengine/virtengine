# Production environment values
# CRITICAL: Review all changes before applying

aws_region         = "us-west-2"
dr_region          = "us-east-1"
vpc_cidr           = "10.2.0.0/16"
kubernetes_version = "1.29"

# SNS topics for alerts (configure with actual ARNs)
alarm_sns_topic_arns = []
