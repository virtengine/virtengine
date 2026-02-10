# TEE Hardware Setup Guide

**Version:** 1.0.0  
**Date:** 2026-02-05  
**Task Reference:** 26E - TEE Hardware Integration and Attestation Testing

---

## Table of Contents

1. [Overview](#overview)
2. [Supported Platforms](#supported-platforms)
3. [Intel SGX Setup](#intel-sgx-setup)
4. [AMD SEV-SNP Setup](#amd-sev-snp-setup)
5. [AWS Nitro Setup](#aws-nitro-setup)
6. [Verification](#verification)
7. [Troubleshooting](#troubleshooting)
8. [CI/CD Configuration](#cicd-configuration)

---

## Overview

VirtEngine supports multiple Trusted Execution Environment (TEE) platforms for secure identity verification and key management. This guide covers hardware setup and configuration for:

- **Intel SGX DCAP** - Process-level isolation with remote attestation
- **AMD SEV-SNP** - VM-level encryption with attestation
- **AWS Nitro Enclaves** - Cloud-native isolated compute

### Platform Selection

| Use Case              | Recommended Platform | Rationale                                 |
| --------------------- | -------------------- | ----------------------------------------- |
| Identity Scoring (ML) | AMD SEV-SNP          | No memory limits, near-native performance |
| Key Management        | Intel SGX            | Smallest TCB, proven sealing              |
| Cloud Deployment      | AWS Nitro            | Managed infrastructure                    |
| Development/Testing   | Simulation           | No hardware required                      |

---

## Supported Platforms

### Intel SGX

| Hardware    | Requirements                                       |
| ----------- | -------------------------------------------------- |
| CPU         | Intel Xeon Scalable (3rd Gen+) or Core (10th Gen+) |
| SGX Version | SGX2 with FLC support preferred                    |
| Driver      | Intel SGX in-kernel driver (Linux 5.11+)           |
| BIOS        | SGX enabled, PRMRR configured                      |

### AMD SEV-SNP

| Hardware   | Requirements                          |
| ---------- | ------------------------------------- |
| CPU        | AMD EPYC Milan (7003) or Genoa (9004) |
| Kernel     | Linux 6.0+ with SNP patches           |
| Hypervisor | KVM with SNP support                  |
| Firmware   | Latest AMD SEV firmware               |

### AWS Nitro

| Environment    | Requirements                             |
| -------------- | ---------------------------------------- |
| Instance Types | c5.xlarge+, m5.xlarge+, r5.xlarge+, etc. |
| AMI            | Amazon Linux 2, AL2023, or compatible    |
| Tools          | nitro-cli, nitro-enclaves-allocator      |
| Kernel         | 5.10+ with Nitro support                 |

---

## Intel SGX Setup

### 1. BIOS Configuration

1. Enter BIOS/UEFI setup
2. Enable Intel SGX (usually under Security or Advanced)
3. Set SGX memory size (PRMRR): Recommended 128MB-256MB
4. Enable Flexible Launch Control (FLC) if available
5. Save and reboot

### 2. Kernel Verification

```bash
# Check kernel version (5.11+ required for in-kernel driver)
uname -r

# Verify SGX support in kernel
grep -i sgx /proc/cpuinfo
dmesg | grep -i sgx
```

### 3. Device Files

```bash
# SGX device files should exist
ls -la /dev/sgx_enclave
ls -la /dev/sgx_provision

# If not present, load the module
sudo modprobe intel_sgx
```

### 4. Install DCAP Components

```bash
# Add Intel SGX repository
echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu jammy main' | \
    sudo tee /etc/apt/sources.list.d/intel-sgx.list

# Add Intel key
wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | \
    sudo apt-key add -

# Install DCAP components
sudo apt update
sudo apt install -y \
    libsgx-dcap-ql \
    libsgx-dcap-default-qpl \
    libsgx-dcap-quote-verify \
    libsgx-urts

# Configure PCCS (optional, for local caching)
sudo apt install -y sgx-dcap-pccs
sudo systemctl enable --now pccs
```

### 5. Permission Configuration

```bash
# Add user to SGX groups
sudo usermod -aG sgx $USER
sudo usermod -aG sgx_prv $USER

# Verify permissions
groups $USER

# Set udev rules (if needed)
echo 'SUBSYSTEM=="misc", KERNEL=="sgx_enclave", MODE="0666"' | \
    sudo tee /etc/udev/rules.d/10-sgx.rules
echo 'SUBSYSTEM=="misc", KERNEL=="sgx_provision", MODE="0660", GROUP="sgx_prv"' | \
    sudo tee -a /etc/udev/rules.d/10-sgx.rules
sudo udevadm control --reload-rules && sudo udevadm trigger
```

### 6. PCCS Configuration

Edit `/etc/sgx_default_qcnl.conf`:

```json
{
  "pccs_url": "https://localhost:8081/sgx/certification/v4/",
  "use_secure_cert": false,
  "collateral_service": "https://api.trustedservices.intel.com/sgx/certification/v4/",
  "pccs_api_version": "3.1",
  "retry_times": 6,
  "retry_delay": 10,
  "local_pck_url": "",
  "pck_cache_expire_hours": 168,
  "verify_collateral_cache_expire_hours": 168,
  "custom_request_options": {
    "get_cert": {
      "timeout": 15000
    }
  }
}
```

---

## AMD SEV-SNP Setup

### 1. Hardware Requirements

- AMD EPYC Milan (7003) or Genoa (9004) processor
- SNP-capable motherboard with latest AGESA firmware
- BIOS with SEV-SNP enabled

### 2. BIOS Configuration

1. Enter BIOS setup
2. Navigate to AMD CBS > CPU Common Options > SEV-SNP
3. Enable SEV, SEV-ES, and SEV-SNP
4. Set minimum ASID: recommended 509+
5. Enable SME if not already enabled
6. Save and reboot

### 3. Kernel Configuration

```bash
# Check kernel version (6.0+ required)
uname -r

# Verify SEV-SNP in kernel config
grep -i sev /boot/config-$(uname -r)
# Should show CONFIG_AMD_MEM_ENCRYPT=y, CONFIG_KVM_AMD_SEV=y

# Check dmesg for SEV initialization
dmesg | grep -i sev
```

### 4. Install SEV Tools

```bash
# Clone AMD SEV tool
git clone https://github.com/AMDESE/sev-tool.git
cd sev-tool
mkdir build && cd build
cmake ..
make -j$(nproc)
sudo make install

# Verify SNP status
sudo sevtool --platform_status
```

### 5. Guest Device Configuration

For SNP guests (inside the confidential VM):

```bash
# Device file should exist
ls -la /dev/sev-guest

# If not present, load the module
sudo modprobe sev-guest

# Verify SNP attestation capability
cat /sys/kernel/debug/sev/snp_enabled
# Should output: 1
```

### 6. Retrieve VCEK Certificate

```bash
# Get chip ID and TCB from attestation report
# The KDS URL format:
# https://kdsintf.amd.com/vcek/v1/{product_name}/{hwid}?blSPL={bl}&teeSPL={tee}&snpSPL={snp}&ucodeSPL={ucode}

# Example for Milan:
curl -o vcek.der "https://kdsintf.amd.com/vcek/v1/Milan/${CHIP_ID}?blSPL=3&teeSPL=0&snpSPL=14&ucodeSPL=209"
```

---

## AWS Nitro Setup

### 1. Launch Nitro-Enabled Instance

```bash
# Create instance with Nitro Enclave support
aws ec2 run-instances \
    --image-id ami-0123456789abcdef0 \
    --instance-type c5.xlarge \
    --enclave-options Enabled=true \
    --key-name my-key \
    --security-group-ids sg-12345678
```

### 2. Install Nitro CLI

```bash
# Amazon Linux 2
sudo amazon-linux-extras install aws-nitro-enclaves-cli -y
sudo yum install aws-nitro-enclaves-cli-devel -y

# Amazon Linux 2023
sudo dnf install aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel -y

# Ubuntu
sudo apt install aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel
```

### 3. Configure Enclave Allocator

Edit `/etc/nitro_enclaves/allocator.yaml`:

```yaml
---
# Enclave memory allocation in MiB
memory_mib: 2048

# Number of CPUs to allocate
cpu_count: 2

# CPU pool (optional)
# cpu_ids:
#   - 2
#   - 3
```

### 4. Start Allocator Service

```bash
# Enable and start the allocator
sudo systemctl enable nitro-enclaves-allocator
sudo systemctl start nitro-enclaves-allocator

# Verify device exists
ls -la /dev/nitro_enclaves

# Add user to ne group
sudo usermod -aG ne $USER
```

### 5. Build and Run Enclave

```bash
# Build enclave image from Docker
nitro-cli build-enclave \
    --docker-uri my-enclave-image:latest \
    --output-file my-enclave.eif

# Run the enclave
nitro-cli run-enclave \
    --eif-path my-enclave.eif \
    --memory 512 \
    --cpu-count 2

# Verify running enclaves
nitro-cli describe-enclaves
```

---

## Verification

### Hardware Detection Test

```bash
# Run VirtEngine hardware detection
./virtengine tee detect

# Expected output:
# TEE Hardware Detection Results:
#   SGX: Available (Version 2, FLC: true)
#   SEV-SNP: Not Available
#   Nitro: Not Available
#   Preferred Platform: SGX
```

### Attestation Test

```bash
# Generate test attestation
./virtengine tee attest --platform auto --output quote.bin

# Verify attestation
./virtengine tee verify --input quote.bin
```

### Integration Test

```bash
# Run TEE integration tests (simulation mode)
go test -v ./pkg/enclave_runtime/... -tags=''

# Run hardware tests (requires hardware)
go test -v ./pkg/enclave_runtime/... -tags='sgx_hardware'
```

---

## Troubleshooting

### SGX Issues

#### "SGX device not found"

```bash
# Check kernel module
lsmod | grep intel_sgx
sudo modprobe intel_sgx

# Check BIOS settings
# Ensure SGX is enabled and PRMRR is set

# Check for conflicts
dmesg | grep -i "sgx\|epc"
```

#### "Permission denied"

```bash
# Add user to groups
sudo usermod -aG sgx,sgx_prv $USER
newgrp sgx

# Check device permissions
ls -la /dev/sgx_*
```

#### "Quote generation failed"

```bash
# Check AESM service
sudo systemctl status aesmd

# Check PCCS connectivity
curl -k https://localhost:8081/sgx/certification/v4/platforms

# Verify PCK certificate availability
sudo /opt/intel/sgx-dcap-pccs/pccs.sh check
```

### SEV-SNP Issues

#### "SEV-SNP device not found"

```bash
# Check if running in SNP guest
cat /sys/kernel/debug/sev/snp_guest_status

# Load guest module
sudo modprobe sev-guest
```

#### "Attestation report failed"

```bash
# Check VMPL level
cat /sys/kernel/debug/sev/vmpl

# Verify SEV-SNP is enabled in hypervisor
virsh domcapabilities | grep -i sev
```

### Nitro Issues

#### "Nitro device not found"

```bash
# Check instance type supports Nitro
aws ec2 describe-instance-types --instance-types c5.xlarge \
    --query 'InstanceTypes[].NitroEnclaveSupport'

# Check enclave option enabled
aws ec2 describe-instances --instance-ids i-xxx \
    --query 'Reservations[].Instances[].EnclaveOptions'
```

#### "Allocator failed"

```bash
# Check allocator service
sudo systemctl status nitro-enclaves-allocator

# Check available memory
cat /proc/meminfo | grep -i nitro

# Restart allocator
sudo systemctl restart nitro-enclaves-allocator
```

---

## CI/CD Configuration

### GitHub Actions

```yaml
name: TEE Tests

on: [push, pull_request]

jobs:
  tee-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Run TEE tests (simulation)
        run: |
          go test -v ./pkg/enclave_runtime/... -tags=''

      # Hardware tests run on self-hosted runners

  tee-hardware-tests:
    runs-on: [self-hosted, sgx]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Run SGX hardware tests
        run: |
          go test -v ./pkg/enclave_runtime/... -tags='sgx_hardware'
```

### Build Tags

| Tag               | Description                           |
| ----------------- | ------------------------------------- |
| (none)            | Simulation mode, no hardware required |
| `sgx_hardware`    | Real SGX SDK calls                    |
| `sev_hardware`    | Real SEV-SNP /dev/sev-guest calls     |
| `nitro_hardware`  | Real Nitro CLI calls                  |
| `e2e.integration` | E2E tests (all platforms)             |

### Test Execution

```bash
# Simulation mode (default)
go test ./pkg/enclave_runtime/...

# SGX hardware mode
go test -tags=sgx_hardware ./pkg/enclave_runtime/sgx/...

# SEV hardware mode
go test -tags=sev_hardware ./pkg/enclave_runtime/sev/...

# All E2E tests
go test -tags=e2e.integration ./tests/e2e/...
```

---

## Security Considerations

### Production Checklist

- [ ] Debug mode disabled in all enclaves
- [ ] Measurement allowlist configured
- [ ] TCB level requirements enforced
- [ ] Certificate pinning enabled
- [ ] Attestation freshness validated
- [ ] Nonce binding verified
- [ ] Sealing key policy set to MRENCLAVE

### Key Management

- Signing keys generated inside enclave only
- Private keys never exported
- Sealed to MRENCLAVE for upgrade resistance
- Key rotation on enclave update

### Audit Logging

Enable TEE audit logging:

```yaml
tee:
  audit:
    enabled: true
    log_attestations: true
    log_key_operations: true
    retention_days: 90
```

---

## References

- [Intel SGX DCAP Documentation](https://download.01.org/intel-sgx/)
- [AMD SEV-SNP ABI Specification](https://www.amd.com/system/files/TechDocs/56860.pdf)
- [AWS Nitro Enclaves User Guide](https://docs.aws.amazon.com/enclaves/latest/user/)
- [VirtEngine TEE Architecture](tee-integration-architecture.md)
- [TEE Security Model](tee-security-model.md)
