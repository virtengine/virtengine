# Staging Environment Variables

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.1.0.0/16"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the EKS public endpoint"
  type        = list(string)
  default     = [] # Configured via tfvars
}
