# VirtEngine RDS Module
# Provides PostgreSQL RDS for chain state backup and application data

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Random Password Generation
# -----------------------------------------------------------------------------
resource "random_password" "master" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# -----------------------------------------------------------------------------
# Secrets Manager for Database Credentials
# -----------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "db_credentials" {
  name                    = "${var.project}-${var.environment}-db-credentials"
  description             = "Database credentials for ${var.project} ${var.environment}"
  recovery_window_in_days = var.environment == "prod" ? 30 : 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = var.master_username
    password = random_password.master.result
    engine   = "postgres"
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    dbname   = var.database_name
  })
}

# -----------------------------------------------------------------------------
# DB Parameter Group
# -----------------------------------------------------------------------------
resource "aws_db_parameter_group" "main" {
  name   = "${var.project}-${var.environment}-pg"
  family = "postgres${split(".", var.engine_version)[0]}"

  parameter {
    name  = "log_min_duration_statement"
    value = var.log_min_duration_statement
  }

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  parameter {
    name  = "log_lock_waits"
    value = "1"
  }

  parameter {
    name  = "log_temp_files"
    value = "0"
  }

  parameter {
    name         = "rds.force_ssl"
    value        = "1"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "shared_preload_libraries"
    value = "pg_stat_statements"
  }

  dynamic "parameter" {
    for_each = var.additional_parameters
    content {
      name         = parameter.value.name
      value        = parameter.value.value
      apply_method = lookup(parameter.value, "apply_method", "immediate")
    }
  }

  tags = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

# -----------------------------------------------------------------------------
# DB Option Group (for enhanced monitoring, etc.)
# -----------------------------------------------------------------------------
resource "aws_db_option_group" "main" {
  name                 = "${var.project}-${var.environment}-og"
  engine_name          = "postgres"
  major_engine_version = split(".", var.engine_version)[0]

  tags = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

# -----------------------------------------------------------------------------
# IAM Role for Enhanced Monitoring
# -----------------------------------------------------------------------------
resource "aws_iam_role" "rds_monitoring" {
  count = var.monitoring_interval > 0 ? 1 : 0
  name  = "${var.project}-${var.environment}-rds-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "monitoring.rds.amazonaws.com"
      }
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  count      = var.monitoring_interval > 0 ? 1 : 0
  role       = aws_iam_role.rds_monitoring[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# -----------------------------------------------------------------------------
# KMS Key for RDS Encryption
# -----------------------------------------------------------------------------
resource "aws_kms_key" "rds" {
  count                   = var.kms_key_arn == "" ? 1 : 0
  description             = "KMS key for RDS ${var.project}-${var.environment}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-rds-kms"
  })
}

resource "aws_kms_alias" "rds" {
  count         = var.kms_key_arn == "" ? 1 : 0
  name          = "alias/${var.project}-${var.environment}-rds"
  target_key_id = aws_kms_key.rds[0].key_id
}

# -----------------------------------------------------------------------------
# RDS Instance
# -----------------------------------------------------------------------------
resource "aws_db_instance" "main" {
  identifier = "${var.project}-${var.environment}"

  # Engine
  engine               = "postgres"
  engine_version       = var.engine_version
  instance_class       = var.instance_class
  parameter_group_name = aws_db_parameter_group.main.name
  option_group_name    = aws_db_option_group.main.name

  # Storage
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type          = var.storage_type
  iops                  = var.storage_type == "io1" ? var.iops : null
  storage_encrypted     = true
  kms_key_id            = var.kms_key_arn != "" ? var.kms_key_arn : aws_kms_key.rds[0].arn

  # Database
  db_name  = var.database_name
  username = var.master_username
  password = random_password.master.result

  # Network
  db_subnet_group_name   = var.db_subnet_group_name
  vpc_security_group_ids = [var.security_group_id]
  publicly_accessible    = false
  port                   = 5432

  # Availability
  multi_az = var.multi_az

  # Backup
  backup_retention_period   = var.backup_retention_period
  backup_window             = var.backup_window
  copy_tags_to_snapshot     = true
  delete_automated_backups  = var.environment != "prod"
  deletion_protection       = var.deletion_protection
  final_snapshot_identifier = var.skip_final_snapshot ? null : "${var.project}-${var.environment}-final-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
  skip_final_snapshot       = var.skip_final_snapshot

  # Maintenance
  maintenance_window          = var.maintenance_window
  auto_minor_version_upgrade  = var.auto_minor_version_upgrade
  allow_major_version_upgrade = false

  # Monitoring
  monitoring_interval                   = var.monitoring_interval
  monitoring_role_arn                   = var.monitoring_interval > 0 ? aws_iam_role.rds_monitoring[0].arn : null
  enabled_cloudwatch_logs_exports       = ["postgresql", "upgrade"]
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null

  # Timeouts
  timeouts {
    create = "60m"
    update = "60m"
    delete = "60m"
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}"
  })

  lifecycle {
    ignore_changes = [
      password,
      final_snapshot_identifier,
    ]
  }
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "cpu" {
  alarm_name          = "${var.project}-${var.environment}-rds-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "RDS CPU utilization above 80%"
  alarm_actions       = var.alarm_actions
  ok_actions          = var.alarm_actions

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.main.id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "storage" {
  alarm_name          = "${var.project}-${var.environment}-rds-storage"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  metric_name         = "FreeStorageSpace"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = var.allocated_storage * 1024 * 1024 * 1024 * 0.2 # 20% free
  alarm_description   = "RDS storage space below 20%"
  alarm_actions       = var.alarm_actions
  ok_actions          = var.alarm_actions

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.main.id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "connections" {
  alarm_name          = "${var.project}-${var.environment}-rds-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = var.max_connections_threshold
  alarm_description   = "RDS connections above threshold"
  alarm_actions       = var.alarm_actions
  ok_actions          = var.alarm_actions

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.main.id
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# Read Replica (optional, for prod)
# -----------------------------------------------------------------------------
resource "aws_db_instance" "replica" {
  count = var.create_read_replica ? 1 : 0

  identifier = "${var.project}-${var.environment}-replica"

  replicate_source_db = aws_db_instance.main.identifier
  instance_class      = var.replica_instance_class != "" ? var.replica_instance_class : var.instance_class

  vpc_security_group_ids = [var.security_group_id]
  publicly_accessible    = false
  port                   = 5432

  storage_encrypted = true
  kms_key_id        = var.kms_key_arn != "" ? var.kms_key_arn : aws_kms_key.rds[0].arn

  auto_minor_version_upgrade = var.auto_minor_version_upgrade

  monitoring_interval = var.monitoring_interval
  monitoring_role_arn = var.monitoring_interval > 0 ? aws_iam_role.rds_monitoring[0].arn : null

  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-replica"
  })
}
