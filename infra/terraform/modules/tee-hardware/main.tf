# VirtEngine TEE Hardware Provisioning Module
# TEE-HW-001: Deploy TEE hardware & attestation in production
#
# This module provisions Trusted Execution Environment (TEE) infrastructure
# across multiple TEE platforms: AWS Nitro, AMD SEV-SNP, and Intel SGX.

terraform {
  required_version = ">= 1.6.0"
}

locals {
  name_prefix = "${var.cluster_name}-tee"

  common_tags = merge(var.tags, {
    Component = "tee-hardware"
    Task      = "TEE-HW-001"
  })

  # Instance types by TEE platform
  nitro_instance_types = [
    "c5.xlarge",      # Nitro-based
    "c5.2xlarge",
    "c5a.xlarge",     # AMD Nitro
    "c6i.xlarge",     # Intel Ice Lake Nitro
    "c6i.2xlarge",
  ]

  sev_snp_instance_types = [
    "c6a.xlarge",     # AMD EPYC Milan (SEV-SNP capable)
    "c6a.2xlarge",
    "c6a.4xlarge",
  ]

  sgx_instance_types = [
    "c5.xlarge",      # SGX capable (limited EPC)
    "c6i.xlarge",     # Intel Ice Lake with SGX
    "c6i.2xlarge",
  ]
}

# =============================================================================
# AWS Nitro Enclave Node Group
# =============================================================================

resource "aws_eks_node_group" "nitro_enclave" {
  count = var.enable_nitro ? 1 : 0

  cluster_name    = var.cluster_name
  node_group_name = "${local.name_prefix}-nitro"
  node_role_arn   = var.node_role_arn
  subnet_ids      = var.subnet_ids

  instance_types = var.nitro_instance_types != null ? var.nitro_instance_types : local.nitro_instance_types
  capacity_type  = "ON_DEMAND"  # Production requires on-demand

  scaling_config {
    desired_size = var.nitro_desired_size
    max_size     = var.nitro_max_size
    min_size     = var.nitro_min_size
  }

  update_config {
    max_unavailable_percentage = 25
  }

  labels = {
    "virtengine.io/tee-platform"   = "nitro"
    "virtengine.io/enclave-ready"  = "true"
    "node.kubernetes.io/instance-type" = "enclave"
  }

  taint {
    key    = "virtengine.io/tee"
    value  = "nitro"
    effect = "NO_SCHEDULE"
  }

  launch_template {
    id      = aws_launch_template.nitro_enclave[0].id
    version = aws_launch_template.nitro_enclave[0].latest_version
  }

  tags = merge(local.common_tags, {
    TEEPlatform = "nitro"
  })

  lifecycle {
    ignore_changes = [scaling_config[0].desired_size]
  }

  depends_on = [aws_launch_template.nitro_enclave]
}

resource "aws_launch_template" "nitro_enclave" {
  count = var.enable_nitro ? 1 : 0

  name_prefix   = "${local.name_prefix}-nitro-"
  description   = "Launch template for Nitro Enclave nodes"

  # Enable Nitro Enclaves
  enclave_options {
    enabled = true
  }

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size           = var.nitro_disk_size
      volume_type           = "gp3"
      encrypted             = true
      kms_key_id            = var.kms_key_arn
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"  # IMDSv2 required
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "enabled"
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(<<-EOF
    #!/bin/bash
    set -ex

    # Install Nitro Enclaves CLI
    amazon-linux-extras enable aws-nitro-enclaves-cli
    yum install -y aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel

    # Add ec2-user to ne group
    usermod -aG ne ec2-user

    # Configure Nitro Enclaves allocator
    cat > /etc/nitro_enclaves/allocator.yaml <<ALLOCATOR
    ---
    memory_mib: ${var.nitro_enclave_memory_mb}
    cpu_count: ${var.nitro_enclave_cpu_count}
    ALLOCATOR

    # Enable and start the allocator
    systemctl enable nitro-enclaves-allocator.service
    systemctl start nitro-enclaves-allocator.service

    # Signal CloudFormation (if applicable)
    /opt/aws/bin/cfn-signal -e $? --stack ${var.cluster_name} --resource NitroASG --region ${var.aws_region} || true
  EOF
  )

  tag_specifications {
    resource_type = "instance"
    tags = merge(local.common_tags, {
      Name        = "${local.name_prefix}-nitro-node"
      TEEPlatform = "nitro"
    })
  }

  tags = local.common_tags
}

# =============================================================================
# AMD SEV-SNP Node Group (Azure compatible, simulated on AWS for multi-cloud)
# =============================================================================

resource "aws_eks_node_group" "sev_snp" {
  count = var.enable_sev_snp ? 1 : 0

  cluster_name    = var.cluster_name
  node_group_name = "${local.name_prefix}-sev-snp"
  node_role_arn   = var.node_role_arn
  subnet_ids      = var.subnet_ids

  instance_types = var.sev_snp_instance_types != null ? var.sev_snp_instance_types : local.sev_snp_instance_types
  capacity_type  = "ON_DEMAND"

  scaling_config {
    desired_size = var.sev_snp_desired_size
    max_size     = var.sev_snp_max_size
    min_size     = var.sev_snp_min_size
  }

  update_config {
    max_unavailable_percentage = 25
  }

  labels = {
    "virtengine.io/tee-platform"  = "sev-snp"
    "virtengine.io/enclave-ready" = "true"
  }

  taint {
    key    = "virtengine.io/tee"
    value  = "sev-snp"
    effect = "NO_SCHEDULE"
  }

  launch_template {
    id      = aws_launch_template.sev_snp[0].id
    version = aws_launch_template.sev_snp[0].latest_version
  }

  tags = merge(local.common_tags, {
    TEEPlatform = "sev-snp"
  })

  lifecycle {
    ignore_changes = [scaling_config[0].desired_size]
  }
}

resource "aws_launch_template" "sev_snp" {
  count = var.enable_sev_snp ? 1 : 0

  name_prefix = "${local.name_prefix}-sev-snp-"
  description = "Launch template for AMD SEV-SNP capable nodes"

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size           = var.sev_snp_disk_size
      volume_type           = "gp3"
      encrypted             = true
      kms_key_id            = var.kms_key_arn
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(<<-EOF
    #!/bin/bash
    set -ex

    # Install SEV-SNP guest tools
    yum install -y kernel-modules-extra

    # Load SEV modules
    modprobe ccp || true
    modprobe kvm_amd || true

    # Verify SEV-SNP availability
    if [ -e /dev/sev-guest ]; then
      echo "SEV-SNP device available"
      chmod 660 /dev/sev-guest
      chown root:kvm /dev/sev-guest
    else
      echo "WARNING: SEV-SNP device not available - running in simulation mode"
    fi

    # Create virtengine user and add to kvm group
    useradd -r -s /sbin/nologin virtengine || true
    usermod -aG kvm virtengine || true
  EOF
  )

  tag_specifications {
    resource_type = "instance"
    tags = merge(local.common_tags, {
      Name        = "${local.name_prefix}-sev-snp-node"
      TEEPlatform = "sev-snp"
    })
  }

  tags = local.common_tags
}

# =============================================================================
# Intel SGX Node Group
# =============================================================================

resource "aws_eks_node_group" "sgx" {
  count = var.enable_sgx ? 1 : 0

  cluster_name    = var.cluster_name
  node_group_name = "${local.name_prefix}-sgx"
  node_role_arn   = var.node_role_arn
  subnet_ids      = var.subnet_ids

  instance_types = var.sgx_instance_types != null ? var.sgx_instance_types : local.sgx_instance_types
  capacity_type  = "ON_DEMAND"

  scaling_config {
    desired_size = var.sgx_desired_size
    max_size     = var.sgx_max_size
    min_size     = var.sgx_min_size
  }

  update_config {
    max_unavailable_percentage = 25
  }

  labels = {
    "virtengine.io/tee-platform"  = "sgx"
    "virtengine.io/enclave-ready" = "true"
  }

  taint {
    key    = "virtengine.io/tee"
    value  = "sgx"
    effect = "NO_SCHEDULE"
  }

  launch_template {
    id      = aws_launch_template.sgx[0].id
    version = aws_launch_template.sgx[0].latest_version
  }

  tags = merge(local.common_tags, {
    TEEPlatform = "sgx"
  })

  lifecycle {
    ignore_changes = [scaling_config[0].desired_size]
  }
}

resource "aws_launch_template" "sgx" {
  count = var.enable_sgx ? 1 : 0

  name_prefix = "${local.name_prefix}-sgx-"
  description = "Launch template for Intel SGX capable nodes"

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size           = var.sgx_disk_size
      volume_type           = "gp3"
      encrypted             = true
      kms_key_id            = var.kms_key_arn
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(<<-EOF
    #!/bin/bash
    set -ex

    # Install Intel SGX driver and SDK
    yum install -y make gcc kernel-devel

    # Install SGX DCAP driver
    # Note: This requires Intel SGX SDK to be available
    wget -O /tmp/sgx_linux_x64_driver.bin https://download.01.org/intel-sgx/latest/linux-latest/distro/rhel8.6-server/sgx_linux_x64_driver_2.11.054c9c4.bin || true
    chmod +x /tmp/sgx_linux_x64_driver.bin
    /tmp/sgx_linux_x64_driver.bin --accept-license || true

    # Install DCAP quote provider
    yum install -y libsgx-dcap-quote-verify libsgx-dcap-default-qpl || true

    # Create SGX device permissions
    if [ -e /dev/sgx/enclave ]; then
      chmod 666 /dev/sgx/enclave
      chmod 666 /dev/sgx/provision || true
    fi

    # Set up PCCS endpoint (Intel Provisioning Certificate Caching Service)
    mkdir -p /etc/sgx_default_qcnl.conf.d
    cat > /etc/sgx_default_qcnl.conf <<QCNL
    {
      "pccs_url": "${var.sgx_pccs_endpoint}",
      "use_secure_cert": true,
      "collateral_service": "https://api.trustedservices.intel.com/sgx/certification/v4/",
      "retry_times": 3,
      "retry_delay": 5
    }
    QCNL
  EOF
  )

  tag_specifications {
    resource_type = "instance"
    tags = merge(local.common_tags, {
      Name        = "${local.name_prefix}-sgx-node"
      TEEPlatform = "sgx"
    })
  }

  tags = local.common_tags
}

# =============================================================================
# Security Group for TEE Attestation Traffic
# =============================================================================

resource "aws_security_group" "tee_attestation" {
  name        = "${local.name_prefix}-attestation-sg"
  description = "Security group for TEE attestation traffic"
  vpc_id      = var.vpc_id

  # Intra-cluster attestation
  ingress {
    description = "TEE attestation protocol (gRPC)"
    from_port   = 8443
    to_port     = 8443
    protocol    = "tcp"
    self        = true
  }

  # Allow PCCS access (SGX)
  egress {
    description = "Intel PCCS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Intel PCCS endpoints
  }

  # Allow KDS access (AMD SEV)
  egress {
    description = "AMD KDS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # AMD KDS endpoints
  }

  # Allow NSM access (Nitro)
  egress {
    description = "AWS KMS for Nitro attestation"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # AWS KMS endpoints
  }

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-attestation-sg"
  })
}

# =============================================================================
# IAM Role for Nitro Enclave Attestation
# =============================================================================

resource "aws_iam_role_policy" "nitro_attestation" {
  count = var.enable_nitro ? 1 : 0

  name = "${local.name_prefix}-nitro-attestation"
  role = var.node_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "NitroEnclaveAttestation"
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey",
        ]
        Resource = var.kms_key_arn
        Condition = {
          StringEqualsIgnoreCase = {
            "kms:RecipientAttestation:ImageSha384" = var.nitro_enclave_image_sha384
          }
        }
      },
      {
        Sid    = "NitroEnclaveDescribe"
        Effect = "Allow"
        Action = [
          "ec2:DescribeInstances",
          "ec2:DescribeVolumes",
        ]
        Resource = "*"
      }
    ]
  })
}

# =============================================================================
# Secrets for TEE Configuration
# =============================================================================

resource "aws_secretsmanager_secret" "tee_config" {
  name        = "${local.name_prefix}/config"
  description = "TEE configuration secrets"
  kms_key_id  = var.kms_key_arn

  tags = local.common_tags
}

resource "aws_secretsmanager_secret_version" "tee_config" {
  secret_id = aws_secretsmanager_secret.tee_config.id

  secret_string = jsonencode({
    sgx_pccs_endpoint        = var.sgx_pccs_endpoint
    amd_kds_base_url         = "https://kdsintf.amd.com/vcek/v1"
    measurement_allowlist    = var.measurement_allowlist
    attestation_cache_ttl    = "5m"
    min_tcb_version          = var.min_tcb_version
    require_hardware         = true
    allow_debug              = false
    production_mode          = true
  })
}

# =============================================================================
# CloudWatch Alarms for TEE Hardware
# =============================================================================

resource "aws_cloudwatch_metric_alarm" "tee_node_health" {
  for_each = toset(compact([
    var.enable_nitro ? "nitro" : "",
    var.enable_sev_snp ? "sev-snp" : "",
    var.enable_sgx ? "sgx" : "",
  ]))

  alarm_name          = "${local.name_prefix}-${each.key}-node-health"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  metric_name         = "cluster_node_count"
  namespace           = "ContainerInsights"
  period              = 300
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "TEE ${each.key} node count has dropped below minimum"
  alarm_actions       = var.alarm_sns_topic_arns

  dimensions = {
    ClusterName = var.cluster_name
    NodeGroup   = "${local.name_prefix}-${each.key}"
  }

  tags = local.common_tags
}
