# Variables for VirtEngine Vault Module

variable "project" {
  description = "Project name for resource naming"
  type        = string
  default     = "virtengine"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the OIDC provider for IRSA"
  type        = string
}

variable "oidc_provider_url" {
  description = "URL of the OIDC provider"
  type        = string
}

variable "vault_namespace" {
  description = "Kubernetes namespace for Vault"
  type        = string
  default     = "vault"
}

variable "vault_chart_version" {
  description = "Version of the Vault Helm chart"
  type        = string
  default     = "0.27.0"
}

variable "vault_version" {
  description = "Version of Vault to deploy"
  type        = string
  default     = "1.15.4"
}

variable "vault_replicas" {
  description = "Number of Vault replicas"
  type        = number
  default     = 3
}

variable "vault_memory_request" {
  description = "Memory request for Vault pods"
  type        = string
  default     = "256Mi"
}

variable "vault_memory_limit" {
  description = "Memory limit for Vault pods"
  type        = string
  default     = "512Mi"
}

variable "vault_cpu_request" {
  description = "CPU request for Vault pods"
  type        = string
  default     = "250m"
}

variable "vault_cpu_limit" {
  description = "CPU limit for Vault pods"
  type        = string
  default     = "500m"
}

variable "storage_class" {
  description = "Storage class for Vault audit logs"
  type        = string
  default     = "gp3"
}

variable "enable_injector" {
  description = "Enable Vault Agent Injector"
  type        = bool
  default     = true
}

variable "enable_csi" {
  description = "Enable Vault CSI Provider"
  type        = bool
  default     = false
}

variable "enable_external_secrets" {
  description = "Enable External Secrets Operator"
  type        = bool
  default     = true
}

variable "external_secrets_chart_version" {
  description = "Version of External Secrets Helm chart"
  type        = string
  default     = "0.9.11"
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
