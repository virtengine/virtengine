# Validator Hardware Requirements

**Version:** 1.0.0  
**Last updated:** 2026-02-06

## Overview
These requirements define the minimum and recommended production hardware
profiles for VirtEngine mainnet validators. They are sized to sustain steady
block production, VEID verification workloads, and on-chain marketplace/HPC
traffic.

## Hardware Profiles

| Profile | CPU | RAM | Storage | Network | Notes |
| --- | --- | --- | --- | --- | --- |
| Minimum (Validator) | 8 cores / 16 threads, x86_64 | 32 GB | 1 TB NVMe (>= 5k IOPS) | 1 Gbps symmetric | Suitable for low-traffic periods; no archive retention |
| Recommended (Validator) | 16 cores / 32 threads, x86_64 | 64 GB | 2 TB NVMe (>= 10k IOPS) | 1–10 Gbps symmetric | Target for mainnet launch + steady-state |
| High-Throughput / Archive | 32+ cores / 64 threads | 128 GB | 4 TB NVMe + cold storage | 10 Gbps symmetric | Archive + analytics + high throughput |

## CPU Requirements
- x86_64 processor with AES-NI support.
- AVX2 recommended for accelerated cryptography.
- No GPU required for validators (VEID scoring is CPU-only and deterministic).

## Storage Requirements
- NVMe SSD required for mainnet validators.
- Sustained write throughput >= 500 MB/s.
- Keep 30% free space for pruning, snapshots, and compaction.
- Archive nodes should use NVMe for the hot dataset and separate HDD or object
  storage for long-term snapshots.

## Network Requirements
- Dedicated 1 Gbps symmetric connection minimum.
- Stable latency < 50 ms to at least 2/3 of the validator set.
- Public P2P port exposed with DDoS protection or upstream filtering.

## OS + Kernel
- Ubuntu 22.04 LTS (or equivalent) recommended.
- Kernel 5.15+.
- NTP enabled and verified (chrony or systemd-timesyncd).

## Security Requirements
- Hardware-backed key storage (HSM or signing service) strongly recommended.
- Separate validator and sentry nodes with strict firewall rules.
- Encrypted disks for validator + state backup volumes.

## Capacity Planning Notes
- VEID verification pipeline runs deterministically on validator CPUs; budget
  2–4 cores for verification during peak onboarding windows.
- Marketplace/HPC activity increases state size; plan for 3–6 months of growth
  per TB on hot NVMe storage.