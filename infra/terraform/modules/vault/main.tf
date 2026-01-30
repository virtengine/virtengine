# VirtEngine Vault Module
# Provides HashiCorp Vault deployment on EKS with AWS Secrets Manager backend

terraform {
  required_version = ">= 1.5.0"
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

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

# -----------------------------------------------------------------------------
# KMS Key for Vault Auto-Unseal
# -----------------------------------------------------------------------------
resource "aws_kms_key" "vault" {
  description             = "KMS key for Vault auto-unseal - ${var.project}-${var.environment}"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  multi_region            = var.environment == "prod"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Vault to use the key"
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_role.vault.arn
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-vault-kms"
  })
}

resource "aws_kms_alias" "vault" {
  name          = "alias/${var.project}-${var.environment}-vault"
  target_key_id = aws_kms_key.vault.key_id
}

# -----------------------------------------------------------------------------
# IAM Role for Vault (IRSA)
# -----------------------------------------------------------------------------
resource "aws_iam_role" "vault" {
  name = "${var.project}-${var.environment}-vault-role"

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
          "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.vault_namespace}:vault"
        }
      }
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "vault_kms" {
  name = "${var.project}-${var.environment}-vault-kms-policy"
  role = aws_iam_role.vault.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = aws_kms_key.vault.arn
      }
    ]
  })
}

resource "aws_iam_role_policy" "vault_secrets" {
  name = "${var.project}-${var.environment}-vault-secrets-policy"
  role = aws_iam_role.vault.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:PutSecretValue",
          "secretsmanager:CreateSecret",
          "secretsmanager:DeleteSecret",
          "secretsmanager:TagResource",
          "secretsmanager:UpdateSecret"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${var.project}/${var.environment}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:ListSecrets"
        ]
        Resource = "*"
      }
    ]
  })
}

# DynamoDB for Vault HA Storage
resource "aws_iam_role_policy" "vault_dynamodb" {
  name = "${var.project}-${var.environment}-vault-dynamodb-policy"
  role = aws_iam_role.vault.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:DescribeLimits",
          "dynamodb:DescribeTimeToLive",
          "dynamodb:ListTagsOfResource",
          "dynamodb:DescribeReservedCapacityOfferings",
          "dynamodb:DescribeReservedCapacity",
          "dynamodb:ListTables",
          "dynamodb:BatchGetItem",
          "dynamodb:BatchWriteItem",
          "dynamodb:CreateTable",
          "dynamodb:DeleteItem",
          "dynamodb:GetItem",
          "dynamodb:GetRecords",
          "dynamodb:PutItem",
          "dynamodb:Query",
          "dynamodb:UpdateItem",
          "dynamodb:Scan",
          "dynamodb:DescribeTable"
        ]
        Resource = [
          aws_dynamodb_table.vault.arn,
          "${aws_dynamodb_table.vault.arn}/*"
        ]
      }
    ]
  })
}

# -----------------------------------------------------------------------------
# DynamoDB Table for Vault HA Storage
# -----------------------------------------------------------------------------
resource "aws_dynamodb_table" "vault" {
  name         = "${var.project}-${var.environment}-vault"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "Path"
  range_key    = "Key"

  attribute {
    name = "Path"
    type = "S"
  }

  attribute {
    name = "Key"
    type = "S"
  }

  point_in_time_recovery {
    enabled = var.environment == "prod"
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.vault.arn
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-vault"
  })
}

# -----------------------------------------------------------------------------
# Kubernetes Namespace
# -----------------------------------------------------------------------------
resource "kubernetes_namespace" "vault" {
  metadata {
    name = var.vault_namespace

    labels = {
      "app.kubernetes.io/name"       = "vault"
      "app.kubernetes.io/instance"   = var.environment
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# -----------------------------------------------------------------------------
# Vault Helm Release
# -----------------------------------------------------------------------------
resource "helm_release" "vault" {
  name       = "vault"
  repository = "https://helm.releases.hashicorp.com"
  chart      = "vault"
  version    = var.vault_chart_version
  namespace  = kubernetes_namespace.vault.metadata[0].name

  values = [
    yamlencode({
      global = {
        enabled = true
      }

      injector = {
        enabled = var.enable_injector
        resources = {
          requests = {
            memory = "256Mi"
            cpu    = "250m"
          }
          limits = {
            memory = "512Mi"
            cpu    = "500m"
          }
        }
      }

      server = {
        enabled = true

        image = {
          repository = "hashicorp/vault"
          tag        = var.vault_version
        }

        resources = {
          requests = {
            memory = var.vault_memory_request
            cpu    = var.vault_cpu_request
          }
          limits = {
            memory = var.vault_memory_limit
            cpu    = var.vault_cpu_limit
          }
        }

        readinessProbe = {
          enabled = true
          path    = "/v1/sys/health?standbyok=true&sealedcode=204&uninitcode=204"
        }

        livenessProbe = {
          enabled     = true
          path        = "/v1/sys/health?standbyok=true"
          initialDelaySeconds = 60
        }

        extraEnvironmentVars = {
          AWS_REGION         = data.aws_region.current.name
          VAULT_SEAL_TYPE    = "awskms"
          VAULT_AWSKMS_SEAL_KEY_ID = aws_kms_key.vault.key_id
        }

        serviceAccount = {
          create = true
          name   = "vault"
          annotations = {
            "eks.amazonaws.com/role-arn" = aws_iam_role.vault.arn
          }
        }

        dataStorage = {
          enabled      = false  # Using DynamoDB
        }

        auditStorage = {
          enabled      = true
          size         = "10Gi"
          storageClass = var.storage_class
        }

        standalone = {
          enabled = false
        }

        ha = {
          enabled  = true
          replicas = var.vault_replicas
          raft = {
            enabled = false  # Using DynamoDB for HA
          }
          config = <<-EOF
            ui = true
            
            listener "tcp" {
              tls_disable = 1
              address = "[::]:8200"
              cluster_address = "[::]:8201"
              telemetry {
                unauthenticated_metrics_access = true
              }
            }
            
            seal "awskms" {
              region     = "${data.aws_region.current.name}"
              kms_key_id = "${aws_kms_key.vault.key_id}"
            }
            
            storage "dynamodb" {
              ha_enabled = "true"
              region     = "${data.aws_region.current.name}"
              table      = "${aws_dynamodb_table.vault.name}"
            }
            
            telemetry {
              prometheus_retention_time = "30s"
              disable_hostname = true
            }
            
            service_registration "kubernetes" {}
          EOF
        }
      }

      ui = {
        enabled         = true
        serviceType     = "ClusterIP"
        serviceNodePort = null
      }

      csi = {
        enabled = var.enable_csi
      }
    })
  ]

  depends_on = [
    kubernetes_namespace.vault,
    aws_dynamodb_table.vault,
    aws_kms_key.vault,
  ]
}

# -----------------------------------------------------------------------------
# External Secrets Operator (for syncing Vault secrets to K8s)
# -----------------------------------------------------------------------------
resource "helm_release" "external_secrets" {
  count = var.enable_external_secrets ? 1 : 0

  name       = "external-secrets"
  repository = "https://charts.external-secrets.io"
  chart      = "external-secrets"
  version    = var.external_secrets_chart_version
  namespace  = "external-secrets"

  create_namespace = true

  values = [
    yamlencode({
      installCRDs = true
      
      serviceAccount = {
        create = true
        name   = "external-secrets"
      }

      resources = {
        requests = {
          cpu    = "100m"
          memory = "128Mi"
        }
        limits = {
          cpu    = "200m"
          memory = "256Mi"
        }
      }

      webhook = {
        port = 9443
      }

      certController = {
        requeueInterval = "5m"
      }
    })
  ]
}

# -----------------------------------------------------------------------------
# Vault SecretStore for External Secrets
# -----------------------------------------------------------------------------
resource "kubernetes_manifest" "vault_secret_store" {
  count = var.enable_external_secrets ? 1 : 0

  manifest = {
    apiVersion = "external-secrets.io/v1beta1"
    kind       = "ClusterSecretStore"
    metadata = {
      name = "vault-backend"
    }
    spec = {
      provider = {
        vault = {
          server  = "http://vault.${var.vault_namespace}.svc.cluster.local:8200"
          path    = "secret"
          version = "v2"
          auth = {
            kubernetes = {
              mountPath = "kubernetes"
              role      = "external-secrets"
              serviceAccountRef = {
                name      = "external-secrets"
                namespace = "external-secrets"
              }
            }
          }
        }
      }
    }
  }

  depends_on = [helm_release.external_secrets, helm_release.vault]
}
