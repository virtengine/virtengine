# Staging environment values
aws_region         = "us-west-2"
vpc_cidr           = "10.1.0.0/16"
kubernetes_version = "1.29"

# Restrict to CI/CD and VPN ranges
allowed_cidr_blocks = [
  "10.0.0.0/8",    # Internal VPN
  "192.168.0.0/16" # Corporate network (example)
]
