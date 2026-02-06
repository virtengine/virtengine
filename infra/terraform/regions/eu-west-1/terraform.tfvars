# EU-West-1 Region Terraform Variable Values

environment        = "prod"
vpc_cidr           = "10.20.0.0/16"
kubernetes_version = "1.29"
validator_count    = 4

cockroachdb_join_addresses = [
  "cockroachdb-0.cockroachdb.cockroachdb.svc.cluster.local:26257",
  "cockroachdb-1.cockroachdb.cockroachdb.svc.cluster.local:26257",
  "cockroachdb-2.cockroachdb.cockroachdb.svc.cluster.local:26257",
]

tags = {
  Project     = "virtengine"
  Environment = "prod"
  Region      = "eu-west-1"
  Role        = "secondary"
  ManagedBy   = "terraform"
}
