# Monitoring and Alerting

This document defines monitoring and alerting practices for production systems.

## Signals

- Metrics: Availability, latency, saturation, and error rates.
- Logs: Security events, authentication, authorization, and system errors.
- Traces: End to end request paths for critical workflows.

## Alerting

- Alerts align to SLOs and documented runbooks.
- Alert thresholds are reviewed after incidents and quarterly.
- Paging is used for high severity events; lower severity uses ticketing.

## Review Cadence

- Weekly: Alert noise review and tuning.
- Monthly: SLO and capacity review.
- Quarterly: Control effectiveness review.

## Evidence

- Alert definitions and routing configuration exports.
- SLO dashboards and monthly review notes.
