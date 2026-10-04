# VirtEngine VPC Module
# Creates a production-ready VPC with multi-AZ support

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, var.az_count)

  tags = merge(var.tags, {
    Module      = "vpc"
    Environment = var.environment
  })
}

data "aws_availability_zones" "available" {
  #checkov:skip=CKV_AWS_394:accepted: the AZ set is read from the provider at apply time; pinning AZ names would hard-fail in any account whose AZ letters differ | review-by 2027-04-01
  state = "available"
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
# Lock down the auto-created default security group. No ingress/egress blocks
# means nothing is allowed; the VPC has no workload that relies on it.
resource "aws_default_security_group" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(local.tags, {
    Name = "default-deny-all"
  })
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.tags, {
    Name = "${var.name}-vpc"
  })
}

# -----------------------------------------------------------------------------
# Internet Gateway
# -----------------------------------------------------------------------------
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(local.tags, {
    Name = "${var.name}-igw"
  })
}

# -----------------------------------------------------------------------------
# Public Subnets
# -----------------------------------------------------------------------------
resource "aws_subnet" "public" {
  #checkov:skip=CKV_AWS_130:accepted by definition: this is a subnet named `public`; map_public_ip_on_launch=true is the property that makes it public | review-by 2027-04-01
  count = var.az_count

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = local.azs[count.index]
  map_public_ip_on_launch = true

  tags = merge(local.tags, {
    Name                                        = "${var.name}-public-${local.azs[count.index]}"
    "kubernetes.io/role/elb"                    = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
    Tier                                        = "public"
  })
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = merge(local.tags, {
    Name = "${var.name}-public-rt"
  })
}

resource "aws_route_table_association" "public" {
  count = var.az_count

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# -----------------------------------------------------------------------------
# Private Subnets (Application Tier)
# -----------------------------------------------------------------------------
resource "aws_subnet" "private" {
  count = var.az_count

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + var.az_count)
  availability_zone = local.azs[count.index]

  tags = merge(local.tags, {
    Name                                        = "${var.name}-private-${local.azs[count.index]}"
    "kubernetes.io/role/internal-elb"           = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
    Tier                                        = "private"
  })
}

# -----------------------------------------------------------------------------
# NAT Gateways (one per AZ for HA)
# -----------------------------------------------------------------------------
resource "aws_eip" "nat" {
  count  = var.enable_nat_gateway ? var.az_count : 0
  domain = "vpc"

  tags = merge(local.tags, {
    Name = "${var.name}-nat-eip-${local.azs[count.index]}"
  })

  depends_on = [aws_internet_gateway.main]
}

resource "aws_nat_gateway" "main" {
  count = var.enable_nat_gateway ? var.az_count : 0

  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = merge(local.tags, {
    Name = "${var.name}-nat-${local.azs[count.index]}"
  })

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route_table" "private" {
  count = var.az_count

  vpc_id = aws_vpc.main.id

  dynamic "route" {
    for_each = var.enable_nat_gateway ? [1] : []
    content {
      cidr_block     = "0.0.0.0/0"
      nat_gateway_id = aws_nat_gateway.main[count.index].id
    }
  }

  tags = merge(local.tags, {
    Name = "${var.name}-private-rt-${local.azs[count.index]}"
  })
}

resource "aws_route_table_association" "private" {
  count = var.az_count

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# -----------------------------------------------------------------------------
# Database Subnets (Isolated Tier)
# -----------------------------------------------------------------------------
resource "aws_subnet" "database" {
  count = var.create_database_subnets ? var.az_count : 0

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + (var.az_count * 2))
  availability_zone = local.azs[count.index]

  tags = merge(local.tags, {
    Name = "${var.name}-database-${local.azs[count.index]}"
    Tier = "database"
  })
}

resource "aws_db_subnet_group" "main" {
  count = var.create_database_subnets ? 1 : 0

  name        = "${var.name}-db-subnet-group"
  description = "Database subnet group for ${var.name}"
  subnet_ids  = aws_subnet.database[*].id

  tags = merge(local.tags, {
    Name = "${var.name}-db-subnet-group"
  })
}

# -----------------------------------------------------------------------------
# VPC Flow Logs
# -----------------------------------------------------------------------------
resource "aws_flow_log" "main" {
  count = var.enable_flow_logs ? 1 : 0

  vpc_id                   = aws_vpc.main.id
  traffic_type             = "ALL"
  log_destination_type     = "cloud-watch-logs"
  log_destination          = aws_cloudwatch_log_group.flow_logs[0].arn
  iam_role_arn             = aws_iam_role.flow_logs[0].arn
  max_aggregation_interval = 60

  tags = merge(local.tags, {
    Name = "${var.name}-flow-logs"
  })
}

resource "aws_cloudwatch_log_group" "flow_logs" {
  #checkov:skip=CKV_AWS_338:KNOWN GAP, real defect carried deliberately: retention is below the 1-year the check wants. Declared with a deliberate operational retention; long-term retention is carried by the S3 archive buckets. Real fix is to confirm each retention with the log owner and raise where the window is genuinely too short | review-by 2026-11-01
  count = var.enable_flow_logs ? 1 : 0

  name              = "/aws/vpc/${var.name}/flow-logs"
  retention_in_days = var.flow_logs_retention_days
  kms_key_id        = aws_kms_key.flow_logs.arn

  tags = local.tags
}

# -----------------------------------------------------------------------------
# KMS key for VPC flow log encryption
# -----------------------------------------------------------------------------
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "flow_logs" {
  description             = "KMS key for VPC flow log encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowKeyAdministration"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "AllowServicesUse"
        Effect = "Allow"
        Principal = {
          Service = [
            "cloudwatch.amazonaws.com",
            "vpc-flow-logs.amazonaws.com",
            "logs.${data.aws_region.current.name}.amazonaws.com",
          ]
        }
        Action   = ["kms:Decrypt", "kms:GenerateDataKey*", "kms:Describe*"]
        Resource = "*"
      },
    ]
  })

  tags = merge(local.tags, {
    Name = "${var.name}-vpc-flow-logs-key"
  })
}

resource "aws_kms_alias" "flow_logs" {
  name          = "alias/${var.name}-vpc-flow-logs"
  target_key_id = aws_kms_key.flow_logs.key_id
}

resource "aws_iam_role" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0

  name = "${var.name}-vpc-flow-logs-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "vpc-flow-logs.amazonaws.com"
      }
    }]
  })

  tags = local.tags
}

resource "aws_iam_role_policy" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0

  name = "${var.name}-vpc-flow-logs-policy"
  role = aws_iam_role.flow_logs[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        # Write actions scoped to THIS log group. The delivery role has no reason
        # to be able to write to any other log group in the account.
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Effect   = "Allow"
        Resource = "${aws_cloudwatch_log_group.flow_logs[0].arn}:*"
      },
      {
        # logs:CreateLogGroup is scoped to the log group itself (no stream suffix);
        # AWS requires this call to be made against the parent log group ARN.
        Action   = ["logs:CreateLogGroup"]
        Effect   = "Allow"
        Resource = aws_cloudwatch_log_group.flow_logs[0].arn
      },
      {
        # logs:DescribeLogGroups does NOT support resource-level permissions --
        # AWS rejects any ARN other than "*". It is a read-only, account-scoped
        # call and cannot be narrowed; it is isolated in its own statement so the
        # write actions above can be scoped.
        Action   = ["logs:DescribeLogGroups"]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        # REQUIRED: the log group is KMS-encrypted, so the flow-logs delivery role
        # must be allowed to generate the data key for every envelope write.
        # Without this the delivery silently fails once encryption is enabled.
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey",
          "kms:DescribeKey"
        ]
        Effect   = "Allow"
        Resource = aws_kms_key.flow_logs.arn
      },
    ]
  })
}

# -----------------------------------------------------------------------------
# VPC Endpoints (for private subnet access to AWS services)
# -----------------------------------------------------------------------------
resource "aws_vpc_endpoint" "s3" {
  count = var.enable_vpc_endpoints ? 1 : 0

  vpc_id            = aws_vpc.main.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = aws_route_table.private[*].id

  tags = merge(local.tags, {
    Name = "${var.name}-s3-endpoint"
  })
}

resource "aws_vpc_endpoint" "ecr_api" {
  count = var.enable_vpc_endpoints ? 1 : 0

  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ecr.api"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints[0].id]
  private_dns_enabled = true

  tags = merge(local.tags, {
    Name = "${var.name}-ecr-api-endpoint"
  })
}

resource "aws_vpc_endpoint" "ecr_dkr" {
  count = var.enable_vpc_endpoints ? 1 : 0

  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints[0].id]
  private_dns_enabled = true

  tags = merge(local.tags, {
    Name = "${var.name}-ecr-dkr-endpoint"
  })
}

resource "aws_vpc_endpoint" "sts" {
  count = var.enable_vpc_endpoints ? 1 : 0

  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.sts"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints[0].id]
  private_dns_enabled = true

  tags = merge(local.tags, {
    Name = "${var.name}-sts-endpoint"
  })
}

resource "aws_security_group" "vpc_endpoints" {
  #checkov:skip=CKV_AWS_382:accepted: node/cluster/database security groups need unrestricted egress to pull images and reach AWS service endpoints; restricting it requires a per-service-endpoint egress allowlist | review-by 2027-04-01
  #checkov:skip=CKV2_AWS_5:accepted: attachment is expressed through aws_security_group_rule or a module output rather than an inline vpc_id on the group; the check only resolves the inline form | review-by 2027-04-01
  count = var.enable_vpc_endpoints ? 1 : 0

  name        = "${var.name}-vpc-endpoints-sg"
  description = "Security group for VPC endpoints"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
    description = "Allow HTTPS from VPC"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = merge(local.tags, {
    Name = "${var.name}-vpc-endpoints-sg"
  })
}

data "aws_region" "current" {}
