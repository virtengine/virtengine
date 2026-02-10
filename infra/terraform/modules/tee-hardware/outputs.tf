# VirtEngine TEE Hardware Module - Outputs
# TEE-HW-001: Deploy TEE hardware & attestation in production

output "nitro_node_group_arn" {
  description = "ARN of the Nitro Enclave node group"
  value       = var.enable_nitro ? aws_eks_node_group.nitro_enclave[0].arn : null
}

output "nitro_node_group_id" {
  description = "ID of the Nitro Enclave node group"
  value       = var.enable_nitro ? aws_eks_node_group.nitro_enclave[0].id : null
}

output "sev_snp_node_group_arn" {
  description = "ARN of the SEV-SNP node group"
  value       = var.enable_sev_snp ? aws_eks_node_group.sev_snp[0].arn : null
}

output "sev_snp_node_group_id" {
  description = "ID of the SEV-SNP node group"
  value       = var.enable_sev_snp ? aws_eks_node_group.sev_snp[0].id : null
}

output "sgx_node_group_arn" {
  description = "ARN of the SGX node group"
  value       = var.enable_sgx ? aws_eks_node_group.sgx[0].arn : null
}

output "sgx_node_group_id" {
  description = "ID of the SGX node group"
  value       = var.enable_sgx ? aws_eks_node_group.sgx[0].id : null
}

output "tee_attestation_security_group_id" {
  description = "Security group ID for TEE attestation traffic"
  value       = aws_security_group.tee_attestation.id
}

output "tee_config_secret_arn" {
  description = "ARN of the TEE configuration secret"
  value       = aws_secretsmanager_secret.tee_config.arn
}

output "enabled_platforms" {
  description = "List of enabled TEE platforms"
  value = compact([
    var.enable_nitro ? "nitro" : "",
    var.enable_sev_snp ? "sev-snp" : "",
    var.enable_sgx ? "sgx" : "",
  ])
}

output "node_labels" {
  description = "Kubernetes labels for TEE node selection"
  value = {
    enclave_ready = "virtengine.io/enclave-ready=true"
    platform = {
      nitro   = "virtengine.io/tee-platform=nitro"
      sev_snp = "virtengine.io/tee-platform=sev-snp"
      sgx     = "virtengine.io/tee-platform=sgx"
    }
  }
}

output "node_taints" {
  description = "Kubernetes taints for TEE nodes"
  value = {
    nitro   = "virtengine.io/tee=nitro:NoSchedule"
    sev_snp = "virtengine.io/tee=sev-snp:NoSchedule"
    sgx     = "virtengine.io/tee=sgx:NoSchedule"
  }
}
