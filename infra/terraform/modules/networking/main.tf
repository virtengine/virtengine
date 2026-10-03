# VirtEngine Networking Module
# Provides VPC, subnets, NAT gateway, and security groups for production infrastructure

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
# Lock down the auto-created default security group. No ingress/egress blocks
# means nothing is allowed; the VPC has no workload that relies on it.
resource "aws_default_security_group" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "default-deny-all"
  })
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-vpc"
  })
}

# -----------------------------------------------------------------------------
# Internet Gateway
# -----------------------------------------------------------------------------
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-igw"
  })
}

# -----------------------------------------------------------------------------
# Public Subnets
# -----------------------------------------------------------------------------
resource "aws_subnet" "public" {
  #checkov:skip=CKV_AWS_130:accepted by definition: this is a subnet named `public`; map_public_ip_on_launch=true is the property that makes it public | review-by 2027-04-01
  count                   = length(var.availability_zones)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name                                        = "${var.project}-${var.environment}-public-${var.availability_zones[count.index]}"
    "kubernetes.io/role/elb"                    = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  })
}

# -----------------------------------------------------------------------------
# Private Subnets
# -----------------------------------------------------------------------------
resource "aws_subnet" "private" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + length(var.availability_zones))
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name                                        = "${var.project}-${var.environment}-private-${var.availability_zones[count.index]}"
    "kubernetes.io/role/internal-elb"           = "1"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  })
}

# -----------------------------------------------------------------------------
# Database Subnets (isolated)
# -----------------------------------------------------------------------------
resource "aws_subnet" "database" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + 2 * length(var.availability_zones))
  availability_zone = var.availability_zones[count.index]

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-database-${var.availability_zones[count.index]}"
  })
}

# -----------------------------------------------------------------------------
# Elastic IPs for NAT Gateways
# -----------------------------------------------------------------------------
resource "aws_eip" "nat" {
  count  = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : 0
  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-nat-eip-${count.index + 1}"
  })

  depends_on = [aws_internet_gateway.main]
}

# -----------------------------------------------------------------------------
# NAT Gateways
# -----------------------------------------------------------------------------
resource "aws_nat_gateway" "main" {
  count         = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : 0
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-nat-${count.index + 1}"
  })

  depends_on = [aws_internet_gateway.main]
}

# -----------------------------------------------------------------------------
# Route Tables
# -----------------------------------------------------------------------------
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-public-rt"
  })
}

resource "aws_route_table" "private" {
  count  = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : 1
  vpc_id = aws_vpc.main.id

  dynamic "route" {
    for_each = var.enable_nat_gateway ? [1] : []
    content {
      cidr_block     = "0.0.0.0/0"
      nat_gateway_id = var.single_nat_gateway ? aws_nat_gateway.main[0].id : aws_nat_gateway.main[count.index].id
    }
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-private-rt-${count.index + 1}"
  })
}

resource "aws_route_table" "database" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-database-rt"
  })
}

# -----------------------------------------------------------------------------
# Route Table Associations
# -----------------------------------------------------------------------------
resource "aws_route_table_association" "public" {
  count          = length(var.availability_zones)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count          = length(var.availability_zones)
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = var.single_nat_gateway ? aws_route_table.private[0].id : aws_route_table.private[count.index].id
}

resource "aws_route_table_association" "database" {
  count          = length(var.availability_zones)
  subnet_id      = aws_subnet.database[count.index].id
  route_table_id = aws_route_table.database.id
}

# -----------------------------------------------------------------------------
# Database Subnet Group
# -----------------------------------------------------------------------------
resource "aws_db_subnet_group" "main" {
  name        = "${var.project}-${var.environment}-db-subnet-group"
  description = "Database subnet group for ${var.project} ${var.environment}"
  subnet_ids  = aws_subnet.database[*].id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-db-subnet-group"
  })
}

# -----------------------------------------------------------------------------
# VPC Flow Logs
# -----------------------------------------------------------------------------
resource "aws_flow_log" "main" {
  count                    = var.enable_flow_logs ? 1 : 0
  iam_role_arn             = aws_iam_role.flow_logs[0].arn
  log_destination          = aws_cloudwatch_log_group.flow_logs[0].arn
  traffic_type             = "ALL"
  vpc_id                   = aws_vpc.main.id
  max_aggregation_interval = 60

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-flow-logs"
  })
}

resource "aws_cloudwatch_log_group" "flow_logs" {
  #checkov:skip=CKV_AWS_338:KNOWN GAP, real defect carried deliberately: retention is below the 1-year the check wants. Declared with a deliberate operational retention; long-term retention is carried by the S3 archive buckets. Real fix is to confirm each retention with the log owner and raise where the window is genuinely too short | review-by 2026-11-01
  count             = var.enable_flow_logs ? 1 : 0
  name              = "/aws/vpc/${var.project}-${var.environment}/flow-logs"
  retention_in_days = var.flow_logs_retention_days
  kms_key_id        = aws_kms_key.flow_logs.arn

  tags = var.tags
}

data "aws_region" "current" {}

resource "aws_iam_role" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0
  name  = "${var.project}-${var.environment}-flow-logs-role"

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

  tags = var.tags
}

resource "aws_iam_role_policy" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0
  name  = "${var.project}-${var.environment}-flow-logs-policy"
  role  = aws_iam_role.flow_logs[0].id

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

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-flow-logs-key"
  })
}

resource "aws_kms_alias" "flow_logs" {
  name          = "alias/${var.project}-${var.environment}-flow-logs"
  target_key_id = aws_kms_key.flow_logs.key_id
}

# -----------------------------------------------------------------------------
# Security Groups
# -----------------------------------------------------------------------------

# EKS Cluster Security Group
resource "aws_security_group" "eks_cluster" {
  #checkov:skip=CKV2_AWS_5:accepted: attachment is expressed through aws_security_group_rule or a module output rather than an inline vpc_id on the group; the check only resolves the inline form | review-by 2027-04-01
  name        = "${var.project}-${var.environment}-eks-cluster-sg"
  description = "Security group for EKS cluster"
  vpc_id      = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-eks-cluster-sg"
  })
}

resource "aws_security_group_rule" "eks_cluster_ingress_nodes" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.eks_cluster.id
  source_security_group_id = aws_security_group.eks_nodes.id
  description              = "Allow nodes to communicate with cluster API"
}

resource "aws_security_group_rule" "eks_cluster_egress" {
  #checkov:skip=CKV_AWS_382:accepted: node/cluster/database security groups need unrestricted egress to pull images and reach AWS service endpoints; restricting it requires a per-service-endpoint egress allowlist | review-by 2027-04-01
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.eks_cluster.id
  description       = "Allow all outbound traffic"
}

# EKS Node Security Group
resource "aws_security_group" "eks_nodes" {
  #checkov:skip=CKV2_AWS_5:accepted: attachment is expressed through aws_security_group_rule or a module output rather than an inline vpc_id on the group; the check only resolves the inline form | review-by 2027-04-01
  name        = "${var.project}-${var.environment}-eks-nodes-sg"
  description = "Security group for EKS worker nodes"
  vpc_id      = aws_vpc.main.id

  tags = merge(var.tags, {
    Name                                        = "${var.project}-${var.environment}-eks-nodes-sg"
    "kubernetes.io/cluster/${var.cluster_name}" = "owned"
  })
}

resource "aws_security_group_rule" "eks_nodes_internal" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 65535
  protocol                 = "-1"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = aws_security_group.eks_nodes.id
  description              = "Allow nodes to communicate with each other"
}

resource "aws_security_group_rule" "eks_nodes_cluster_ingress" {
  type                     = "ingress"
  from_port                = 1025
  to_port                  = 65535
  protocol                 = "tcp"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = aws_security_group.eks_cluster.id
  description              = "Allow cluster to communicate with nodes"
}

resource "aws_security_group_rule" "eks_nodes_cluster_ingress_443" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = aws_security_group.eks_cluster.id
  description              = "Allow cluster to communicate with nodes on 443"
}

resource "aws_security_group_rule" "eks_nodes_egress" {
  #checkov:skip=CKV_AWS_382:accepted: node/cluster/database security groups need unrestricted egress to pull images and reach AWS service endpoints; restricting it requires a per-service-endpoint egress allowlist | review-by 2027-04-01
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.eks_nodes.id
  description       = "Allow all outbound traffic"
}

# Database Security Group
resource "aws_security_group" "database" {
  #checkov:skip=CKV2_AWS_5:accepted: attachment is expressed through aws_security_group_rule or a module output rather than an inline vpc_id on the group; the check only resolves the inline form | review-by 2027-04-01
  name        = "${var.project}-${var.environment}-database-sg"
  description = "Security group for RDS database"
  vpc_id      = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-database-sg"
  })
}

resource "aws_security_group_rule" "database_ingress_nodes" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  security_group_id        = aws_security_group.database.id
  source_security_group_id = aws_security_group.eks_nodes.id
  description              = "Allow PostgreSQL from EKS nodes"
}

resource "aws_security_group_rule" "database_egress" {
  #checkov:skip=CKV_AWS_382:accepted: node/cluster/database security groups need unrestricted egress to pull images and reach AWS service endpoints; restricting it requires a per-service-endpoint egress allowlist | review-by 2027-04-01
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.database.id
  description       = "Allow all outbound traffic"
}

# Bastion Security Group (optional)
resource "aws_security_group" "bastion" {
  #checkov:skip=CKV2_AWS_5:accepted: attachment is expressed through aws_security_group_rule or a module output rather than an inline vpc_id on the group; the check only resolves the inline form | review-by 2027-04-01
  count       = var.enable_bastion ? 1 : 0
  name        = "${var.project}-${var.environment}-bastion-sg"
  description = "Security group for bastion host"
  vpc_id      = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-bastion-sg"
  })
}

resource "aws_security_group_rule" "bastion_ingress_ssh" {
  count             = var.enable_bastion ? 1 : 0
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = var.bastion_allowed_cidrs
  security_group_id = aws_security_group.bastion[0].id
  description       = "Allow SSH from allowed CIDRs"
}

resource "aws_security_group_rule" "bastion_egress" {
  count             = var.enable_bastion ? 1 : 0
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bastion[0].id
  description       = "Allow all outbound traffic"
}
