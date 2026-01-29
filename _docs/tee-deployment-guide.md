# VirtEngine TEE Deployment Guide

## Overview

This guide covers deploying Trusted Execution Environment (TEE) infrastructure for VirtEngine validators. TEE enables secure identity verification by processing sensitive biometric and document data inside hardware-isolated enclaves.

VirtEngine's VEID (Verified Identity) system requires TEE for:
- **Facial verification** - Comparing live selfie against document photo
- **Liveness detection** - Ensuring real person presence (not photo/video attack)
- **Document OCR** - Extracting PII from identity documents
- **ML scoring** - Computing identity confidence scores deterministically

All sensitive operations occur inside enclaves where data is encrypted in memory and inaccessible to the host operating system, hypervisor, or cloud provider.

---

## Table of Contents

1. [TEE Platform Comparison](#tee-platform-comparison)
2. [Hardware Requirements](#hardware-requirements)
3. [Intel SGX Deployment](#intel-sgx-deployment)
4. [AMD SEV-SNP Deployment](#amd-sev-snp-deployment)
5. [AWS Nitro Deployment](#aws-nitro-deployment)
6. [Enclave Manager Configuration](#enclave-manager-configuration)
7. [Attestation Setup](#attestation-setup)
8. [Monitoring and Operations](#monitoring-and-operations)
9. [Troubleshooting](#troubleshooting)
10. [Security Considerations](#security-considerations)
11. [Production Checklist](#production-checklist)

---

## TEE Platform Comparison

| Feature | Intel SGX | AMD SEV-SNP | AWS Nitro |
|---------|-----------|-------------|-----------|
| Isolation | Enclave (process) | VM-level | VM-level |
| Memory Protected | 128MB-512MB (PRM) | Full VM | Full VM |
| Cloud Support | Limited | Azure, GCP | AWS Only |
| On-Prem Support | Yes | Yes | No |
| Attestation | DCAP, EPID | VCEK, ASK | NSM |
| Ease of Deployment | Medium | Medium | Easy |
| Recommended For | High-security | Private cloud | AWS validators |

### Platform Selection Guidance

```
┌─────────────────────────────────────────────────────────────────┐
│                    TEE Platform Decision Tree                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Running on AWS?                                                 │
│       │                                                          │
│       ├── YES ──► Use AWS Nitro (easiest, best integration)     │
│       │                                                          │
│       └── NO ──► Running on-premises?                           │
│                       │                                          │
│                       ├── YES ──► Intel SGX (most secure)       │
│                       │           or AMD SEV-SNP (more memory)  │
│                       │                                          │
│                       └── NO ──► Check cloud provider:          │
│                                   • Azure → AMD SEV-SNP         │
│                                   • GCP → AMD SEV-SNP           │
│                                   • Other → Intel SGX           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Hardware Requirements

### Intel SGX

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | Intel Xeon E3 v6 | Intel Xeon Scalable 3rd gen+ |
| EPC Memory | 64MB | 256MB+ |
| System RAM | 16GB | 32GB+ |
| Storage | 100GB SSD | 500GB NVMe |
| Network | 1 Gbps | 10 Gbps |

**BIOS Requirements:**
- SGX enabled (not "Software Controlled")
- Flexible Launch Control (FLC) enabled
- Memory encryption enabled

**Verify SGX Support:**
```bash
# Check CPU support
cpuid | grep -i sgx

# Expected output:
#    SGX: Software Guard Extensions supported = true
#    SGX_LC: SGX launch configuration supported = true

# Check kernel support
dmesg | grep -i sgx
# Expected: sgx: EPC section ... initialized
```

### AMD SEV-SNP

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | AMD EPYC 7003 (Milan) | AMD EPYC 9004 (Genoa) |
| System RAM | 32GB | 64GB+ |
| Storage | 200GB SSD | 1TB NVMe |
| Network | 1 Gbps | 10 Gbps |

**Firmware Requirements:**
- SEV-SNP capable firmware (version 1.51+)
- PSP firmware updated to latest

**Verify SNP Support:**
```bash
# Check CPU support
dmesg | grep -i sev

# Expected output:
# SEV-SNP supported: yes
# SEV-ES INIT://platform status

# Check device availability
ls -la /dev/sev*
# Expected: /dev/sev-guest
```

### AWS Nitro

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| Instance Type | c5.xlarge | c5.4xlarge or m5.4xlarge |
| vCPUs | 4 | 16 |
| Memory | 8GB | 64GB |
| Enclave Memory | 2GB | 8GB |
| EBS | 100GB gp3 | 500GB gp3 |

**Instance Requirements:**
- Nitro Enclave-enabled instance family (c5, m5, r5, c6i, m6i, etc.)
- NOT supported: t3, t2, graviton (ARM), mac instances

**Verify Nitro Support:**
```bash
# Check if running on Nitro
aws ec2 describe-instance-types --instance-types $(curl -s http://169.254.169.254/latest/meta-data/instance-type) \
  --query "InstanceTypes[0].EnclaveOptions.Supported"

# Expected: true
```

---

## Intel SGX Deployment

### Step 1: Install SGX Driver and SDK

```bash
# Add Intel SGX repository
echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu focal main' | \
  sudo tee /etc/apt/sources.list.d/intel-sgx.list

wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | \
  sudo apt-key add -

sudo apt-get update

# Install SGX packages
sudo apt-get install -y \
  libsgx-epid \
  libsgx-quote-ex \
  libsgx-dcap-ql \
  libsgx-dcap-default-qpl \
  libsgx-urts \
  sgx-aesm-service \
  libsgx-aesm-quote-ex-plugin \
  libsgx-aesm-ecdsa-plugin

# Start AESM service
sudo systemctl enable aesmd
sudo systemctl start aesmd

# Verify AESM is running
sudo systemctl status aesmd
```

### Step 2: Configure PCCS for DCAP Attestation

The Provisioning Certificate Caching Service (PCCS) caches Intel attestation collateral locally.

```bash
# Install PCCS
sudo apt-get install -y sgx-dcap-pccs

# Configure PCCS
sudo /opt/intel/sgx-dcap-pccs/install.sh
```

**PCCS Configuration (`/opt/intel/sgx-dcap-pccs/config/default.json`):**
```json
{
  "HTTPS_PORT": 8081,
  "hosts": "127.0.0.1",
  "uri": "https://api.trustedservices.intel.com/sgx/certification/v4/",
  "ApiKey": "YOUR_INTEL_PCS_API_KEY",
  "proxy": "",
  "RefreshSchedule": "0 0 1 * * *",
  "CachingFillMode": "LAZY",
  "AdminTokenHash": "GENERATE_SECURE_HASH",
  "UserTokenHash": "GENERATE_SECURE_HASH"
}
```

> **Note:** Get your Intel PCS API key from https://api.portal.trustedservices.intel.com/

```bash
# Start PCCS
sudo systemctl enable pccs
sudo systemctl start pccs

# Verify PCCS health
curl -k https://localhost:8081/sgx/certification/v4/rootcacrl
```

### Step 3: Build and Sign Enclave Binary

```bash
# Clone VirtEngine enclave source
cd /opt/virtengine
git clone https://github.com/virtengine/veid-enclave.git
cd veid-enclave

# Install Gramine (enclave framework)
sudo curl -fsSL https://packages.gramineproject.io/gramine-keyring.gpg | \
  sudo gpg --dearmor -o /usr/share/keyrings/gramine-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/gramine-keyring.gpg] https://packages.gramineproject.io/ focal main" | \
  sudo tee /etc/apt/sources.list.d/gramine.list
sudo apt-get update
sudo apt-get install -y gramine

# Generate signing key (production: use HSM-backed key)
gramine-sgx-gen-private-key enclave-key.pem

# Build enclave
make SGX=1 SGX_SIGN_KEY=enclave-key.pem

# Output: veid_scoring.manifest.sgx, veid_scoring.sig
```

### Step 4: Run Enclave with VirtEngine Daemon

**Enclave Service Configuration (`/etc/systemd/system/veid-enclave.service`):**
```ini
[Unit]
Description=VirtEngine VEID SGX Enclave
After=network.target aesmd.service
Requires=aesmd.service

[Service]
Type=simple
User=virtengine
Group=virtengine
WorkingDirectory=/opt/virtengine/veid-enclave
ExecStart=/usr/bin/gramine-sgx ./veid_scoring
Restart=always
RestartSec=5
Environment=ENCLAVE_MODE=production
Environment=GRPC_LISTEN_ADDR=unix:///var/run/virtengine/enclave.sock

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/run/virtengine

[Install]
WantedBy=multi-user.target
```

```bash
# Create runtime directory
sudo mkdir -p /var/run/virtengine
sudo chown virtengine:virtengine /var/run/virtengine

# Enable and start enclave service
sudo systemctl daemon-reload
sudo systemctl enable veid-enclave
sudo systemctl start veid-enclave

# Verify enclave is running
sudo systemctl status veid-enclave
gramine-sgx-get-token --sig veid_scoring.sig
```

### Step 5: Register Enclave Measurement On-Chain

```bash
# Get enclave MRENCLAVE (measurement)
MRENCLAVE=$(gramine-sgx-sigstruct-view veid_scoring.sig | grep mr_enclave | awk '{print $2}')
echo "MRENCLAVE: $MRENCLAVE"

# Get MRSIGNER (signer identity)
MRSIGNER=$(gramine-sgx-sigstruct-view veid_scoring.sig | grep mr_signer | awk '{print $2}')
echo "MRSIGNER: $MRSIGNER"

# Register measurement on-chain (requires validator key)
virtengine tx veid register-enclave \
  --platform sgx \
  --mrenclave $MRENCLAVE \
  --mrsigner $MRSIGNER \
  --min-isvsvn 1 \
  --from validator \
  --chain-id virtengine-1 \
  --gas auto \
  --gas-adjustment 1.5 \
  -y
```

---

## AMD SEV-SNP Deployment

### Step 1: Verify SNP Support

```bash
# Check CPU capabilities
sudo dmesg | grep -i "SEV-SNP"

# Verify kernel support (Linux 6.0+)
uname -r

# Check SEV device
ls -la /dev/sev*
# Expected: /dev/sev, /dev/sev-guest

# Check SNP is enabled in firmware
sudo sevctl ok
```

### Step 2: Configure SNP-Enabled Kernel

```bash
# Install SNP-enabled kernel (Ubuntu 22.04+)
sudo apt-get install -y linux-image-generic-hwe-22.04

# Add kernel parameters
sudo sed -i 's/GRUB_CMDLINE_LINUX_DEFAULT="/GRUB_CMDLINE_LINUX_DEFAULT="mem_encrypt=on kvm_amd.sev=1 /' /etc/default/grub
sudo update-grub
sudo reboot

# After reboot, verify
dmesg | grep -i "Memory Encryption"
```

### Step 3: Install SNP Tools

```bash
# Install sev-guest tools
git clone https://github.com/AMDESE/sev-guest.git
cd sev-guest
make
sudo make install

# Install snpguest for attestation
cargo install snpguest

# Verify installation
snpguest --version
```

### Step 4: Create Confidential VM

**VM Configuration (`/etc/virtengine/snp-vm.xml`):**
```xml
<domain type='kvm'>
  <name>veid-enclave</name>
  <memory unit='GiB'>8</memory>
  <vcpu>4</vcpu>
  <os>
    <type arch='x86_64' machine='q35'>hvm</type>
    <loader readonly='yes' type='pflash'>/usr/share/OVMF/OVMF_CODE.fd</loader>
    <nvram>/var/lib/libvirt/qemu/nvram/veid-enclave_VARS.fd</nvram>
    <boot dev='hd'/>
  </os>
  <features>
    <acpi/>
    <apic/>
  </features>
  <cpu mode='host-passthrough'>
    <feature policy='require' name='sev'/>
  </cpu>
  <launchSecurity type='sev-snp'>
    <policy>0x30000</policy>
    <guestVisibleWorkarounds/>
    <idBlock/>
    <idAuth/>
    <hostData/>
  </launchSecurity>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/var/lib/libvirt/images/veid-enclave.qcow2'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
  </devices>
</domain>
```

```bash
# Create VM disk
sudo qemu-img create -f qcow2 /var/lib/libvirt/images/veid-enclave.qcow2 50G

# Install OS into VM (use your preferred method)
sudo virt-install \
  --name veid-enclave-temp \
  --ram 8192 \
  --vcpus 4 \
  --disk path=/var/lib/libvirt/images/veid-enclave.qcow2,format=qcow2 \
  --os-variant ubuntu22.04 \
  --cdrom /path/to/ubuntu-22.04-server.iso \
  --network network=default \
  --graphics vnc

# After OS install, convert to SNP VM
virsh define /etc/virtengine/snp-vm.xml
virsh start veid-enclave
```

### Step 5: Set Up Attestation with AMD KDS

```bash
# Inside the SNP guest VM
# Get attestation report
snpguest report attestation.bin request.bin --random

# Fetch certificate chain from AMD KDS
snpguest fetch vcek der attestation.bin certs/

# Verify attestation locally
snpguest verify attestation attestation.bin certs/

# Extract launch measurement
MEASUREMENT=$(snpguest display report attestation.bin | grep "Measurement:" | awk '{print $2}')
echo "Launch Measurement: $MEASUREMENT"
```

### Step 6: Register Launch Measurement On-Chain

```bash
# From the host (or any machine with virtengine CLI)
virtengine tx veid register-enclave \
  --platform sev-snp \
  --measurement $MEASUREMENT \
  --policy 0x30000 \
  --min-version 1 \
  --from validator \
  --chain-id virtengine-1 \
  --gas auto \
  --gas-adjustment 1.5 \
  -y

# Query registered enclaves
virtengine query veid registered-enclaves --platform sev-snp
```

---

## AWS Nitro Deployment

### Step 1: Launch Enclave-Enabled Instance

```bash
# Create instance with enclave support (via AWS CLI)
aws ec2 run-instances \
  --image-id ami-0abcdef1234567890 \
  --instance-type c5.xlarge \
  --enclave-options 'Enabled=true' \
  --key-name my-key-pair \
  --security-group-ids sg-0123456789abcdef0 \
  --subnet-id subnet-0123456789abcdef0 \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=virtengine-tee}]'
```

**Or via Terraform:**
```hcl
resource "aws_instance" "virtengine_tee" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "c5.xlarge"

  enclave_options {
    enabled = true
  }

  tags = {
    Name = "virtengine-tee"
  }
}
```

### Step 2: Install Nitro CLI

```bash
# SSH into instance
ssh -i my-key-pair.pem ec2-user@<instance-ip>

# Install Nitro Enclaves CLI (Amazon Linux 2)
sudo amazon-linux-extras install aws-nitro-enclaves-cli -y
sudo yum install aws-nitro-enclaves-cli-devel -y

# Add user to nitro group
sudo usermod -aG ne ec2-user

# Allocate resources for enclaves
sudo tee /etc/nitro_enclaves/allocator.yaml << EOF
---
memory_mib: 4096
cpu_count: 2
EOF

# Start allocator service
sudo systemctl enable nitro-enclaves-allocator
sudo systemctl start nitro-enclaves-allocator

# Verify installation
nitro-cli --version
```

### Step 3: Build Enclave Image (EIF)

```bash
# Clone VirtEngine enclave source
git clone https://github.com/virtengine/veid-enclave.git
cd veid-enclave

# Build Docker image for enclave
docker build -t veid-enclave:latest -f Dockerfile.nitro .

# Build Nitro Enclave Image File (EIF)
nitro-cli build-enclave \
  --docker-uri veid-enclave:latest \
  --output-file veid_scoring.eif

# Output example:
# Start building the Enclave Image...
# Enclave Image successfully created.
# {
#   "Measurements": {
#     "HashAlgorithm": "Sha384 { ... }",
#     "PCR0": "abc123...",  <-- Enclave image measurement
#     "PCR1": "def456...",  <-- Linux kernel and bootstrap
#     "PCR2": "ghi789..."   <-- Application
#   }
# }

# Save measurements for on-chain registration
nitro-cli build-enclave --docker-uri veid-enclave:latest --output-file veid_scoring.eif 2>&1 | \
  grep -A 4 "PCR0" > /opt/virtengine/enclave/measurements.json
```

### Step 4: Configure vsock Communication

The enclave communicates with the parent instance via vsock (virtual socket).

**Enclave Application Configuration:**
```yaml
# /opt/virtengine/enclave/config.yaml
server:
  listen: "vsock://3:5000"  # CID 3 = enclave, port 5000
  max_connections: 100

parent:
  cid: 3                    # Parent instance CID
  health_port: 5001

attestation:
  include_user_data: true
  nonce_source: "parent"

logging:
  level: "info"
  format: "json"
```

**Parent Instance Proxy (`/etc/virtengine/vsock-proxy.yaml`):**
```yaml
proxy:
  listen: "unix:///var/run/virtengine/enclave.sock"
  enclave_cid: 16           # Assigned when enclave starts
  enclave_port: 5000
  
  timeout: 30s
  max_message_size: 10485760  # 10MB

health_check:
  enabled: true
  interval: 10s
  port: 5001
```

### Step 5: Run Enclave as Systemd Service

**Enclave Service (`/etc/systemd/system/veid-enclave.service`):**
```ini
[Unit]
Description=VirtEngine VEID Nitro Enclave
After=network.target nitro-enclaves-allocator.service
Requires=nitro-enclaves-allocator.service

[Service]
Type=simple
User=ec2-user
Group=ne
ExecStartPre=/usr/bin/nitro-cli terminate-enclave --all || true
ExecStart=/usr/bin/nitro-cli run-enclave \
  --eif-path /opt/virtengine/enclave/veid_scoring.eif \
  --cpu-count 2 \
  --memory 2048 \
  --enclave-cid 16 \
  --debug-mode
ExecStop=/usr/bin/nitro-cli terminate-enclave --all
Restart=always
RestartSec=10

# Environment
Environment=NITRO_CLI_BLOBS=/opt/virtengine/enclave

[Install]
WantedBy=multi-user.target
```

**vsock Proxy Service (`/etc/systemd/system/veid-vsock-proxy.service`):**
```ini
[Unit]
Description=VirtEngine VEID vsock Proxy
After=veid-enclave.service
Requires=veid-enclave.service

[Service]
Type=simple
User=virtengine
ExecStart=/opt/virtengine/bin/vsock-proxy \
  --config /etc/virtengine/vsock-proxy.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable veid-enclave veid-vsock-proxy
sudo systemctl start veid-enclave veid-vsock-proxy

# Verify enclave is running
nitro-cli describe-enclaves

# Expected output:
# [
#   {
#     "EnclaveName": "veid_scoring",
#     "EnclaveID": "i-abc123-enc123",
#     "ProcessID": 12345,
#     "EnclaveCID": 16,
#     "NumberOfCPUs": 2,
#     "CPUIDs": [1, 3],
#     "MemoryMiB": 2048,
#     "State": "RUNNING"
#   }
# ]
```

### Step 6: Register PCR Measurements On-Chain

```bash
# Get enclave PCR values
PCR0=$(nitro-cli describe-enclaves | jq -r '.[0].Measurements.PCR0')
PCR1=$(nitro-cli describe-enclaves | jq -r '.[0].Measurements.PCR1')
PCR2=$(nitro-cli describe-enclaves | jq -r '.[0].Measurements.PCR2')

# Register on-chain
virtengine tx veid register-enclave \
  --platform nitro \
  --pcr0 $PCR0 \
  --pcr1 $PCR1 \
  --pcr2 $PCR2 \
  --from validator \
  --chain-id virtengine-1 \
  --gas auto \
  --gas-adjustment 1.5 \
  -y
```

---

## Enclave Manager Configuration

The Enclave Manager provides a unified interface across TEE platforms with automatic failover.

### Configuration File

**`/etc/virtengine/enclave-manager.yaml`:**
```yaml
# Enclave Manager Configuration
# Manages multiple TEE backends with priority-based selection

manager:
  # Selection strategy: priority, round-robin, least-loaded
  selection_strategy: priority
  
  # Health check configuration
  health_check_interval: 30s
  health_check_timeout: 10s
  
  # Failover settings
  enable_failover: true
  max_retries: 3
  retry_delay: 5s
  
  # Circuit breaker
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    reset_timeout: 60s

# TEE backend configurations
backends:
  - id: sgx-primary
    type: sgx
    priority: 1
    enabled: true
    config:
      enclave_path: /opt/virtengine/enclave/veid_scoring.manifest.sgx
      socket_path: /var/run/virtengine/sgx-enclave.sock
      dcap_pccs_url: https://localhost:8081
      timeout: 30s
      max_concurrent: 10

  - id: nitro-backup
    type: nitro
    priority: 2
    enabled: true
    config:
      enclave_image: /opt/virtengine/enclave/veid_scoring.eif
      enclave_cid: 16
      enclave_port: 5000
      cpu_count: 2
      memory_mb: 2048
      debug_mode: false

  - id: sev-snp-tertiary
    type: sev-snp
    priority: 3
    enabled: false
    config:
      vm_name: veid-enclave
      socket_path: /var/run/virtengine/snp-enclave.sock
      attestation_url: https://kdsintf.amd.com

# Attestation configuration
attestation:
  # Require fresh attestation for each request
  require_fresh: false
  
  # Cache attestation reports
  cache_ttl: 3600s
  
  # Verification settings
  verification:
    check_revocation: true
    allow_debug: false
    min_tcb_level: "UpToDate"

# Metrics and observability
metrics:
  enabled: true
  listen: ":9090"
  path: "/metrics"

logging:
  level: info
  format: json
  output: /var/log/virtengine/enclave-manager.log
```

### Systemd Service

**`/etc/systemd/system/enclave-manager.service`:**
```ini
[Unit]
Description=VirtEngine Enclave Manager
After=network.target
Wants=veid-enclave.service

[Service]
Type=simple
User=virtengine
Group=virtengine
ExecStart=/opt/virtengine/bin/enclave-manager \
  --config /etc/virtengine/enclave-manager.yaml
Restart=always
RestartSec=5

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/run/virtengine /var/log/virtengine

[Install]
WantedBy=multi-user.target
```

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        VirtEngine Validator Node                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐     ┌─────────────────────────────────────────────┐   │
│  │  virtengine  │     │            Enclave Manager                   │   │
│  │    daemon    │────▶│  ┌─────────┐ ┌─────────┐ ┌─────────────┐    │   │
│  │              │     │  │ Health  │ │ Circuit │ │   Request   │    │   │
│  └──────────────┘     │  │ Checker │ │ Breaker │ │   Router    │    │   │
│         │             │  └────┬────┘ └────┬────┘ └──────┬──────┘    │   │
│         │             │       │           │             │            │   │
│         ▼             │  ┌────▼───────────▼─────────────▼────────┐  │   │
│  ┌──────────────┐     │  │              Backend Pool              │  │   │
│  │  Prometheus  │◀────│  │                                        │  │   │
│  │   Metrics    │     │  │  Priority 1    Priority 2    Priority 3│  │   │
│  └──────────────┘     │  │  ┌───────┐    ┌───────┐    ┌─────────┐ │  │   │
│                       │  │  │  SGX  │    │ Nitro │    │ SEV-SNP │ │  │   │
│                       │  │  └───┬───┘    └───┬───┘    └────┬────┘ │  │   │
│                       │  └──────┼────────────┼─────────────┼──────┘  │   │
│                       └─────────┼────────────┼─────────────┼─────────┘   │
│                                 │            │             │             │
├─────────────────────────────────┼────────────┼─────────────┼─────────────┤
│         Hardware/Cloud          │            │             │             │
│                                 ▼            ▼             ▼             │
│  ┌─────────────────────┐  ┌───────────┐  ┌─────────────────────┐        │
│  │    Intel SGX        │  │  AWS      │  │    AMD SEV-SNP      │        │
│  │    Enclave          │  │  Nitro    │  │    Confidential VM  │        │
│  │  (Gramine runtime)  │  │  Enclave  │  │                     │        │
│  └─────────────────────┘  └───────────┘  └─────────────────────┘        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Attestation Setup

### Intel SGX DCAP Attestation

```bash
# Verify DCAP quote generation
cd /opt/virtengine/veid-enclave
./test_quote_generation

# Register quote provider
sudo cp /etc/sgx_default_qcnl.conf.template /etc/sgx_default_qcnl.conf

# Configure quote provider
sudo tee /etc/sgx_default_qcnl.conf << EOF
{
  "pccs_url": "https://localhost:8081/sgx/certification/v4/",
  "use_secure_cert": false,
  "collateral_service": "https://api.trustedservices.intel.com/sgx/certification/v4/",
  "retry_times": 6,
  "retry_delay": 10,
  "pck_cache_expire_hours": 168,
  "verify_collateral_cache_expire_hours": 168,
  "local_pck_url": ""
}
EOF

# Test attestation
virtengine enclave test-attestation --platform sgx
```

### AMD SEV-SNP Attestation

```bash
# Inside SNP guest: Generate attestation report
snpguest report attestation.bin request.bin --random

# Fetch VCEK certificate from AMD KDS
snpguest fetch vcek der attestation.bin ./certs/

# Verify locally
snpguest verify attestation attestation.bin ./certs/

# Test from VirtEngine
virtengine enclave test-attestation --platform sev-snp
```

### AWS Nitro Attestation

```bash
# Get attestation document from enclave
nitro-cli describe-enclaves | jq -r '.[0].Measurements'

# Verify via AWS Nitro SDK (inside enclave)
# The enclave uses aws-nitro-enclaves-nsm-api for attestation

# Test from VirtEngine
virtengine enclave test-attestation --platform nitro
```

### On-Chain Attestation Verification

```bash
# Query attestation status
virtengine query veid enclave-attestation \
  --validator $(virtengine keys show validator -a) \
  --output json | jq

# Submit fresh attestation
virtengine tx veid submit-attestation \
  --platform sgx \
  --attestation-report /path/to/report.bin \
  --from validator \
  --chain-id virtengine-1 \
  -y
```

---

## Monitoring and Operations

### Prometheus Metrics

The Enclave Manager exposes metrics at `:9090/metrics`:

```
# HELP enclave_requests_total Total enclave requests
# TYPE enclave_requests_total counter
enclave_requests_total{backend="sgx-primary",status="success"} 12345
enclave_requests_total{backend="sgx-primary",status="failure"} 23

# HELP enclave_request_duration_seconds Request duration
# TYPE enclave_request_duration_seconds histogram
enclave_request_duration_seconds_bucket{backend="sgx-primary",le="0.1"} 10000
enclave_request_duration_seconds_bucket{backend="sgx-primary",le="0.5"} 12000
enclave_request_duration_seconds_bucket{backend="sgx-primary",le="1.0"} 12300

# HELP enclave_health_status Backend health (1=healthy, 0=unhealthy)
# TYPE enclave_health_status gauge
enclave_health_status{backend="sgx-primary"} 1
enclave_health_status{backend="nitro-backup"} 1

# HELP enclave_memory_usage_bytes Enclave memory usage
# TYPE enclave_memory_usage_bytes gauge
enclave_memory_usage_bytes{backend="sgx-primary"} 134217728

# HELP enclave_attestation_age_seconds Time since last attestation
# TYPE enclave_attestation_age_seconds gauge
enclave_attestation_age_seconds{backend="sgx-primary"} 1800
```

### Grafana Dashboard

Import the VirtEngine TEE dashboard from `/opt/virtengine/monitoring/dashboards/tee-monitoring.json`:

```json
{
  "dashboard": {
    "title": "VirtEngine TEE Monitoring",
    "panels": [
      {
        "title": "Enclave Health",
        "type": "stat",
        "targets": [
          {
            "expr": "enclave_health_status",
            "legendFormat": "{{backend}}"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(enclave_requests_total[5m])",
            "legendFormat": "{{backend}} - {{status}}"
          }
        ]
      },
      {
        "title": "Request Latency (p99)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(enclave_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{backend}}"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "enclave_memory_usage_bytes",
            "legendFormat": "{{backend}}"
          }
        ]
      }
    ]
  }
}
```

### Alert Rules

**`/etc/prometheus/rules/tee-alerts.yml`:**
```yaml
groups:
  - name: tee-alerts
    rules:
      - alert: EnclaveUnhealthy
        expr: enclave_health_status == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "TEE enclave {{ $labels.backend }} is unhealthy"
          description: "Enclave has been unhealthy for more than 5 minutes"

      - alert: EnclaveHighLatency
        expr: histogram_quantile(0.99, rate(enclave_request_duration_seconds_bucket[5m])) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High latency on enclave {{ $labels.backend }}"
          description: "p99 latency is {{ $value }}s, exceeding 5s threshold"

      - alert: EnclaveHighErrorRate
        expr: |
          rate(enclave_requests_total{status="failure"}[5m]) /
          rate(enclave_requests_total[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on enclave {{ $labels.backend }}"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: AttestationExpiring
        expr: enclave_attestation_age_seconds > 82800  # 23 hours
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Attestation expiring soon for {{ $labels.backend }}"
          description: "Attestation is {{ $value | humanizeDuration }} old"

      - alert: AllEnclavesDown
        expr: sum(enclave_health_status) == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "All TEE enclaves are down"
          description: "No healthy enclave backends available"
```

### Log Aggregation

**Fluentd Configuration (`/etc/fluent/fluent.conf`):**
```
<source>
  @type tail
  path /var/log/virtengine/enclave-manager.log
  pos_file /var/log/fluent/enclave-manager.pos
  tag virtengine.enclave
  <parse>
    @type json
    time_key timestamp
    time_format %Y-%m-%dT%H:%M:%S.%NZ
  </parse>
</source>

<filter virtengine.enclave>
  @type record_transformer
  <record>
    hostname "#{Socket.gethostname}"
    service enclave-manager
  </record>
</filter>

<match virtengine.**>
  @type elasticsearch
  host elasticsearch.internal
  port 9200
  index_name virtengine-logs
  type_name _doc
</match>
```

---

## Troubleshooting

### Intel SGX Issues

#### Issue: "SGX device not found"
```bash
# Check if SGX is enabled in BIOS
cpuid | grep -i sgx

# If not showing SGX support:
# 1. Enter BIOS setup
# 2. Enable SGX (may be under Security or Advanced CPU settings)
# 3. Set SGX mode to "Enabled" (not "Software Controlled")
# 4. Enable Flexible Launch Control (FLC)
# 5. Save and reboot

# Load SGX driver
sudo modprobe intel_sgx

# Verify device exists
ls -la /dev/sgx*
```

#### Issue: "AESM service failed to start"
```bash
# Check AESM logs
sudo journalctl -u aesmd -f

# Common fix: reinstall AESM
sudo apt-get remove --purge sgx-aesm-service
sudo apt-get install sgx-aesm-service

# Restart
sudo systemctl restart aesmd
```

#### Issue: "Quote generation failed - PCCS unreachable"
```bash
# Test PCCS connectivity
curl -k https://localhost:8081/sgx/certification/v4/rootcacrl

# Check PCCS logs
sudo journalctl -u pccs -f

# Verify API key is valid
curl -X GET "https://api.trustedservices.intel.com/sgx/certification/v4/pckcrl?ca=platform" \
  -H "Ocp-Apim-Subscription-Key: YOUR_API_KEY"
```

#### Issue: "Enclave signature verification failed"
```bash
# Check enclave signature
gramine-sgx-sigstruct-view veid_scoring.sig

# Verify MRSIGNER matches registered value
virtengine query veid registered-enclaves --platform sgx

# Re-sign if key mismatch
gramine-sgx-sign --manifest veid_scoring.manifest --output veid_scoring.manifest.sgx
```

### AMD SEV-SNP Issues

#### Issue: "SNP not available"
```bash
# Check firmware version
sudo dmesg | grep -i "sev-snp\|sev firmware"

# Update firmware if needed
sudo apt-get install amd-sev-snp-firmware
sudo modprobe ccp

# Verify SNP is enabled
cat /sys/module/kvm_amd/parameters/sev_snp
# Should output: Y
```

#### Issue: "Failed to launch SNP guest"
```bash
# Check available SNP ASIDs
cat /sys/kernel/debug/sev/snp_asid_count

# Verify OVMF supports SNP
strings /usr/share/OVMF/OVMF_CODE.fd | grep -i sev

# Check libvirt logs
sudo journalctl -u libvirtd -f
```

#### Issue: "Attestation verification failed"
```bash
# Fetch fresh certificates
snpguest fetch vcek der attestation.bin ./certs/ --force

# Check certificate chain
openssl verify -CAfile ./certs/ark.pem -untrusted ./certs/ask.pem ./certs/vcek.pem

# Verify report format
snpguest display report attestation.bin
```

### AWS Nitro Issues

#### Issue: "Enclave failed to start - insufficient resources"
```bash
# Check allocated resources
cat /etc/nitro_enclaves/allocator.yaml

# Reduce enclave memory/CPU or increase allocation
sudo systemctl stop nitro-enclaves-allocator

sudo tee /etc/nitro_enclaves/allocator.yaml << EOF
memory_mib: 8192
cpu_count: 4
EOF

sudo systemctl start nitro-enclaves-allocator

# Verify allocation
nitro-cli describe-enclaves
```

#### Issue: "vsock connection refused"
```bash
# Check enclave is running
nitro-cli describe-enclaves

# Verify CID assignment
ENCLAVE_CID=$(nitro-cli describe-enclaves | jq -r '.[0].EnclaveCID')
echo "Enclave CID: $ENCLAVE_CID"

# Test vsock connection
socat - VSOCK-CONNECT:$ENCLAVE_CID:5000

# Check proxy configuration
cat /etc/virtengine/vsock-proxy.yaml
```

#### Issue: "Attestation document invalid"
```bash
# Get fresh attestation from enclave
nitro-cli describe-enclaves | jq '.[] | .Measurements'

# Verify PCR values match registered
virtengine query veid registered-enclaves --platform nitro

# If mismatch, re-register or rebuild enclave
nitro-cli build-enclave --docker-uri veid-enclave:latest --output-file new-veid.eif
```

### General Issues

#### Issue: "Enclave Manager failover not working"
```bash
# Check circuit breaker status
curl http://localhost:9090/metrics | grep circuit_breaker

# Reset circuit breaker
virtengine enclave reset-circuit-breaker --backend sgx-primary

# Check health of all backends
virtengine enclave health --all
```

#### Issue: "High latency on enclave requests"
```bash
# Check enclave memory pressure
# For SGX:
cat /sys/kernel/debug/x86/sgx_encl_nr

# For Nitro:
nitro-cli describe-enclaves | jq '.[0].MemoryMiB'

# Profile request handling
virtengine enclave profile --duration 60s --output profile.json
```

---

## Security Considerations

### Measurement Allowlist Management

**Only allow production-signed enclaves:**
```bash
# View current allowlist
virtengine query veid enclave-allowlist

# Add new measurement (governance proposal)
virtengine tx gov submit-proposal add-enclave-measurement \
  --platform sgx \
  --mrenclave "abc123..." \
  --mrsigner "def456..." \
  --description "VEID Scoring Enclave v1.2.0" \
  --deposit 1000000uve \
  --from validator \
  -y

# Remove compromised measurement (governance proposal)
virtengine tx gov submit-proposal remove-enclave-measurement \
  --platform sgx \
  --mrenclave "old-measurement..." \
  --reason "Security vulnerability CVE-2025-XXXX" \
  --from validator \
  -y
```

### Key Rotation Procedures

**Rotate enclave signing key:**
```bash
# Generate new signing key (HSM recommended)
gramine-sgx-gen-private-key new-enclave-key.pem

# Build enclave with new key
make SGX=1 SGX_SIGN_KEY=new-enclave-key.pem

# Register new measurement BEFORE deploying
virtengine tx veid register-enclave \
  --platform sgx \
  --mrenclave $(gramine-sgx-sigstruct-view veid_scoring.sig | grep mr_enclave | awk '{print $2}') \
  --mrsigner $(gramine-sgx-sigstruct-view veid_scoring.sig | grep mr_signer | awk '{print $2}') \
  --from validator \
  -y

# Wait for on-chain confirmation
virtengine query tx <tx-hash>

# Deploy new enclave
sudo systemctl restart veid-enclave

# After successful rotation, remove old measurement via governance
```

### Debug Mode Restrictions

> **⚠️ WARNING:** Never run debug-mode enclaves in production!

Debug mode enclaves:
- Allow memory inspection by host
- Do not provide confidentiality guarantees
- Are rejected by on-chain attestation verification

```bash
# Verify enclave is NOT in debug mode
gramine-sgx-sigstruct-view veid_scoring.sig | grep -i debug
# Should show: debug = false

# For Nitro, omit --debug-mode flag
nitro-cli run-enclave --eif-path veid_scoring.eif --cpu-count 2 --memory 2048
# Do NOT add: --debug-mode
```

### Incident Response for Enclave Compromise

**If enclave key or measurement is compromised:**

1. **Immediate Actions (within 1 hour):**
   ```bash
   # Notify other validators via emergency channel
   # Submit emergency governance proposal to remove measurement
   virtengine tx gov submit-proposal emergency-remove-enclave \
     --platform sgx \
     --mrenclave "compromised-measurement" \
     --from validator \
     --emergency \
     -y
   
   # Stop local enclave
   sudo systemctl stop veid-enclave enclave-manager
   ```

2. **Short-term (within 24 hours):**
   ```bash
   # Generate new signing key
   # Rebuild enclave with security patches
   # Test in isolated environment
   # Register new measurement
   ```

3. **Long-term:**
   - Conduct post-incident review
   - Update key management procedures
   - Implement additional monitoring

### Network Security

```bash
# Enclave should only communicate via:
# 1. Unix socket to enclave manager (local)
# 2. vsock to parent instance (Nitro only)

# Block external network access for enclave processes
sudo iptables -A OUTPUT -m owner --uid-owner virtengine -j DROP
sudo iptables -A OUTPUT -m owner --uid-owner virtengine -d 127.0.0.1 -j ACCEPT
```

---

## Production Checklist

### Pre-Deployment

- [ ] **Hardware verified** - TEE platform requirements met
- [ ] **BIOS configured** - SGX/SEV enabled with correct settings
- [ ] **Kernel updated** - Required kernel version installed
- [ ] **Drivers installed** - SGX/SEV drivers loaded successfully
- [ ] **Attestation tested** - Quote/report generation works
- [ ] **Network configured** - Firewall rules in place

### Enclave Setup

- [ ] **Enclave built** - Production (non-debug) build completed
- [ ] **Signature verified** - Enclave signed with production key
- [ ] **Measurement recorded** - MRENCLAVE/PCR values documented
- [ ] **On-chain registered** - Measurement registered and confirmed
- [ ] **Systemd configured** - Services enabled and tested
- [ ] **Failover tested** - Multi-backend failover verified

### Security

- [ ] **Debug mode disabled** - No debug enclaves in production
- [ ] **Keys secured** - Signing keys stored in HSM/secure storage
- [ ] **Access restricted** - Minimal permissions for enclave user
- [ ] **Audit logging enabled** - All enclave operations logged
- [ ] **Secrets rotated** - Initial secrets replaced with production values

### Monitoring

- [ ] **Metrics exported** - Prometheus scraping enclave metrics
- [ ] **Dashboards deployed** - Grafana dashboards configured
- [ ] **Alerts configured** - Critical alerts tested and routed
- [ ] **Logs aggregated** - Centralized logging operational
- [ ] **On-call configured** - Runbooks and escalation paths defined

### Operations

- [ ] **Backup procedures** - Enclave configuration backed up
- [ ] **Recovery tested** - Disaster recovery procedure verified
- [ ] **Documentation updated** - Runbooks reflect current setup
- [ ] **Team trained** - Operations team familiar with TEE
- [ ] **Incident response** - Compromise response plan documented

### Final Verification

```bash
# Run comprehensive health check
virtengine enclave health-check --comprehensive

# Expected output:
# ✓ SGX enclave: healthy
# ✓ Attestation: valid (expires in 23h)
# ✓ Measurement: registered on-chain
# ✓ Failover: tested successfully
# ✓ Metrics: exporting normally
# ✓ All checks passed!
```

---

## Additional Resources

### Vendor Documentation

- **Intel SGX:** https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html
- **Intel DCAP:** https://download.01.org/intel-sgx/sgx-dcap/1.19/linux/docs/
- **AMD SEV-SNP:** https://www.amd.com/en/developer/sev.html
- **AWS Nitro Enclaves:** https://docs.aws.amazon.com/enclaves/latest/user/

### VirtEngine Resources

- **Architecture:** [tee-integration-architecture.md](tee-integration-architecture.md)
- **Security Model:** [tee-security-model.md](tee-security-model.md)
- **Migration Plan:** [tee-migration-plan.md](tee-migration-plan.md)
- **Threat Model:** [threat-model.md](threat-model.md)

### Community

- **Discord:** #validators-tee channel
- **Forum:** https://forum.virtengine.io/c/validators
- **GitHub Issues:** https://github.com/virtengine/virtengine/issues

---

*Last updated: January 2026*
*Document version: 1.0.0*
