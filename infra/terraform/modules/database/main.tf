# VirtEngine CockroachDB Multi-Region Database Module
# Provisions CockroachDB across multiple regions using Helm

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }
}

locals {
  tags = merge(var.tags, {
    Module      = "database"
    Environment = var.environment
  })

  namespace = "cockroachdb"
}

# -----------------------------------------------------------------------------
# Kubernetes Namespace
# -----------------------------------------------------------------------------
resource "kubernetes_namespace" "cockroachdb" {
  metadata {
    name = local.namespace

    labels = {
      "app.kubernetes.io/name"       = "cockroachdb"
      "app.kubernetes.io/managed-by" = "terraform"
      "virtengine.io/region"         = var.region
    }
  }
}

# -----------------------------------------------------------------------------
# KMS Key for Database Encryption at Rest
# -----------------------------------------------------------------------------
resource "aws_kms_key" "cockroachdb" {
  description             = "KMS key for CockroachDB encryption at rest in ${var.region}"
  deletion_window_in_days = 14
  enable_key_rotation     = true

  tags = merge(local.tags, {
    Name = "${var.cluster_name}-cockroachdb-key"
  })
}

resource "aws_kms_alias" "cockroachdb" {
  name          = "alias/${var.cluster_name}-cockroachdb"
  target_key_id = aws_kms_key.cockroachdb.key_id
}

# -----------------------------------------------------------------------------
# S3 Bucket for Backups
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "backups" {
  bucket = "virtengine-cockroachdb-backup-${var.region}"

  tags = merge(local.tags, {
    Name = "virtengine-cockroachdb-backup-${var.region}"
  })
}

resource "aws_s3_bucket_versioning" "backups" {
  bucket = aws_s3_bucket.backups.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.cockroachdb.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "backup-retention"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = var.backup_retention_days
    }
  }
}

resource "aws_s3_bucket_public_access_block" "backups" {
  bucket = aws_s3_bucket.backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# -----------------------------------------------------------------------------
# IAM Role for CockroachDB Backup (IRSA)
# -----------------------------------------------------------------------------
resource "aws_iam_role" "cockroachdb_backup" {
  name = "${var.cluster_name}-cockroachdb-backup-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRoleWithWebIdentity"
      Effect = "Allow"
      Principal = {
        Federated = var.oidc_provider_arn
      }
      Condition = {
        StringEquals = {
          "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${local.namespace}:cockroachdb"
        }
      }
    }]
  })

  tags = local.tags
}

resource "aws_iam_role_policy" "cockroachdb_backup" {
  name = "${var.cluster_name}-cockroachdb-backup-policy"
  role = aws_iam_role.cockroachdb_backup.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:DeleteObject",
          "s3:GetBucketLocation",
        ]
        Resource = [
          aws_s3_bucket.backups.arn,
          "${aws_s3_bucket.backups.arn}/*",
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:GenerateDataKey",
        ]
        Resource = [aws_kms_key.cockroachdb.arn]
      },
    ]
  })
}

# -----------------------------------------------------------------------------
# CockroachDB Helm Release
# -----------------------------------------------------------------------------
resource "helm_release" "cockroachdb" {
  name       = "cockroachdb"
  repository = "https://charts.cockroachdb.com/"
  chart      = "cockroachdb"
  version    = var.cockroachdb_chart_version
  namespace  = kubernetes_namespace.cockroachdb.metadata[0].name

  values = [
    yamlencode({
      statefulset = {
        replicas = var.replicas
      }

      conf = {
        locality       = "region=${var.region}"
        single-node    = false
        join           = var.join_addresses
        cache          = var.cache_size
        max-sql-memory = var.max_sql_memory
      }

      tls = {
        enabled = true
      }

      storage = {
        persistentVolume = {
          size         = var.storage_size
          storageClass = var.storage_class
        }
      }

      resources = {
        requests = {
          cpu    = var.cpu_request
          memory = var.memory_request
        }
        limits = {
          cpu    = var.cpu_limit
          memory = var.memory_limit
        }
      }

      nodeSelector = {
        "virtengine.io/region" = var.region
      }

      tolerations = var.tolerations

      serviceAccount = {
        annotations = {
          "eks.amazonaws.com/role-arn" = aws_iam_role.cockroachdb_backup.arn
        }
      }

      labels = {
        "virtengine.io/component" = "database"
        "virtengine.io/region"    = var.region
      }
    })
  ]

  depends_on = [kubernetes_namespace.cockroachdb]
}

# -----------------------------------------------------------------------------
# Backup CronJob
# -----------------------------------------------------------------------------
resource "kubernetes_cron_job_v1" "backup" {
  metadata {
    name      = "cockroachdb-backup"
    namespace = local.namespace
  }

  spec {
    schedule = var.backup_schedule

    job_template {
      metadata {}
      spec {
        template {
          metadata {}
          spec {
            service_account_name = "cockroachdb"
            restart_policy       = "OnFailure"

            container {
              name  = "backup"
              image = "cockroachdb/cockroach:${var.cockroachdb_version}"

              command = [
                "/bin/bash",
                "-c",
                <<-EOT
                cockroach sql --certs-dir=/cockroach/cockroach-certs \
                  --host=cockroachdb-public \
                  -e "BACKUP INTO 's3://virtengine-cockroachdb-backup-${var.region}/backups?AUTH=implicit' WITH revision_history;"
                EOT
              ]

              volume_mount {
                name       = "certs"
                mount_path = "/cockroach/cockroach-certs"
                read_only  = true
              }
            }

            volume {
              name = "certs"
              secret {
                secret_name  = "cockroachdb-client-secret"
                default_mode = "0400"
              }
            }
          }
        }
      }
    }
  }

  depends_on = [helm_release.cockroachdb]
}

# -----------------------------------------------------------------------------
# CloudWatch Alarms for Database Health
# -----------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "backup_age" {
  alarm_name          = "${var.cluster_name}-cockroachdb-backup-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "backup_age_seconds"
  namespace           = "VirtEngine/CockroachDB"
  period              = 3600
  statistic           = "Maximum"
  threshold           = var.backup_max_age_seconds
  alarm_description   = "CockroachDB backup age exceeds ${var.backup_max_age_seconds}s in ${var.region}"
  alarm_actions       = var.alarm_sns_topic_arns

  dimensions = {
    Region = var.region
  }

  tags = local.tags
}
