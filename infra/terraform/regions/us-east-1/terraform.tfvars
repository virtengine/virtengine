# US-East-1 Region Terraform Variable Values

environment        = "prod"
vpc_cidr           = "10.10.0.0/16"
kubernetes_version = "1.29"
validator_count    = 4

cockroachdb_join_addresses = [
  "cockroachdb-0.cockroachdb.cockroachdb.svc.cluster.local:26257",
  "cockroachdb-1.cockroachdb.cockroachdb.svc.cluster.local:26257",
  "cockroachdb-2.cockroachdb.cockroachdb.svc.cluster.local:26257",
]

enable_cross_region_peering = false

tags = {
  Project     = "virtengine"
  Environment = "prod"
  Region      = "us-east-1"
  Role        = "primary"
  ManagedBy   = "terraform"
}
