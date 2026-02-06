# VirtEngine ML Security Audit Scope

## Overview
Security assessment of machine learning models used in VEID identity
verification, focusing on adversarial robustness, bias, and secure deployment.

## Objectives
- Evaluate robustness against adversarial inputs and model extraction.
- Measure bias and fairness across demographic groups.
- Validate secure deployment controls (encryption, auth, logging).
- Confirm deterministic inference requirements are met.

## In-Scope Models

### Facial Verification Model
- Architecture review and training assumptions.
- Adversarial input testing and evasion resilience.
- Bias assessment across demographics and lighting conditions.
- Presentation attack detection (spoofing, masks).

### Liveness Detection Model
- Spoofing and replay resistance.
- Injection attack testing (image/video substitution).
- Timing-based or sequence-based liveness checks.

### OCR / Document Extraction Model
- Document forgery detection and tamper resistance.
- Data extraction accuracy and error handling.
- Adversarial document testing (altered fields, synthetic docs).

## Test Categories

### Adversarial Robustness
- Evasion attacks (FGSM, PGD, CW).
- Poisoning attacks on training data (if training pipeline available).
- Model extraction and inversion attempts.
- Membership inference and privacy leakage.

### Bias and Fairness
- Demographic parity metrics.
- Equalized odds and subgroup accuracy.
- False positive/negative rates by group.
- Intersectional analysis where data allows.

### Deployment Security
- Model encryption at rest and in transit.
- Inference API auth, rate limiting, and audit logging.
- Input validation and normalization.
- Output sanitization to avoid leaking model internals.

## Data and Environment Requirements
- Approved evaluation dataset access with anonymization.
- Secure test environment with audit logging enabled.
- Deterministic inference configuration for reproducibility.

## Out of Scope
- Data collection agreements and legal review.
- Model training pipeline performance tuning.
- UI-based identity verification flows.

## Deliverables
- Adversarial robustness report.
- Bias assessment report and fairness metrics.
- Security recommendations with remediation priorities.
- Updated model cards and risk disclosures.

## Timeline
- Estimated duration: 3-4 weeks.
- Draft report by week 3.
- Final report and retest verification by week 4.

## Points of Contact
- ML Security Lead: ml-security@virtengine.io
- Data Governance: privacy@virtengine.io
- Audit Coordinator: audit@virtengine.io
