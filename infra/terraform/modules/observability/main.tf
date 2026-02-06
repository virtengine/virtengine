# VirtEngine Multi-Region Observability Module
# Prometheus federation, centralized logging, cross-region alerting

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }
}

locals {
  tags = merge(var.tags, {
    Module      = "observability"
    Environment = var.environment
  })

  namespace = "monitoring"
}

# -----------------------------------------------------------------------------
# Kubernetes Namespace
# -----------------------------------------------------------------------------
resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = local.namespace

    labels = {
      "app.kubernetes.io/name"       = "monitoring"
      "app.kubernetes.io/managed-by" = "terraform"
      "virtengine.io/region"         = var.region
    }
  }
}

# -----------------------------------------------------------------------------
# Prometheus with Federation Support
# -----------------------------------------------------------------------------
resource "helm_release" "prometheus_stack" {
  name       = "kube-prometheus-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = var.prometheus_stack_version
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  values = [
    yamlencode({
      prometheus = {
        prometheusSpec = {
          retention     = var.prometheus_retention
          retentionSize = var.prometheus_retention_size
          externalLabels = {
            cluster = var.cluster_name
            region  = var.region
          }

          storageSpec = {
            volumeClaimTemplate = {
              spec = {
                storageClassName = var.storage_class
                accessModes      = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = var.prometheus_storage_size
                  }
                }
              }
            }
          }

          resources = {
            requests = {
              cpu    = "500m"
              memory = "2Gi"
            }
            limits = {
              cpu    = "2"
              memory = "4Gi"
            }
          }

          serviceMonitorSelectorNilUsesHelmValues = false
          podMonitorSelectorNilUsesHelmValues     = false

          # Federation: add remote-write targets for central Prometheus
          remoteWrite = var.is_primary_region ? [] : [
            {
              url = var.central_prometheus_url
              writeRelabelConfigs = [
                {
                  sourceLabels = ["__name__"]
                  regex        = "up|virtengine_.*|cockroachdb_.*|node_.*|container_.*"
                  action       = "keep"
                }
              ]
            }
          ]

          # Federation: accept scrapes from other regions
          additionalScrapeConfigs = var.is_primary_region ? [
            for r in var.federation_targets : {
              job_name        = "federation-${r.region}"
              scrape_interval = "30s"
              honor_labels    = true
              metrics_path    = "/federate"
              params = {
                "match[]" = [
                  "{job=~\".+\"}",
                ]
              }
              static_configs = [
                {
                  targets = [r.prometheus_endpoint]
                  labels = {
                    federated_region = r.region
                  }
                }
              ]
            }
          ] : []
        }
      }

      alertmanager = {
        alertmanagerSpec = {
          storage = {
            volumeClaimTemplate = {
              spec = {
                storageClassName = var.storage_class
                accessModes      = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = "10Gi"
                  }
                }
              }
            }
          }
        }

        config = {
          global = {
            resolve_timeout = "5m"
          }
          route = {
            group_by        = ["alertname", "cluster", "region"]
            group_wait      = "30s"
            group_interval  = "5m"
            repeat_interval = "4h"
            receiver        = "sns"
            routes = [
              {
                matchers        = ["severity = critical"]
                receiver        = "sns"
                repeat_interval = "1h"
              },
              {
                matchers        = ["alertname = RegionDown"]
                receiver        = "pagerduty"
                repeat_interval = "5m"
              },
            ]
          }
          receivers = [
            {
              name = "sns"
              sns_configs = [
                {
                  topic_arn = var.alert_sns_topic_arn
                  sigv4 = {
                    region = var.region
                  }
                  send_resolved = true
                }
              ]
            },
            {
              name            = "pagerduty"
              pagerduty_configs = var.pagerduty_service_key != "" ? [
                {
                  service_key = var.pagerduty_service_key
                  send_resolved = true
                }
              ] : []
            },
          ]
        }
      }

      grafana = {
        enabled = var.is_primary_region

        persistence = {
          enabled          = true
          storageClassName = var.storage_class
          size             = "10Gi"
        }

        sidecar = {
          dashboards = {
            enabled = true
            label   = "grafana_dashboard"
          }
          datasources = {
            enabled = true
            label   = "grafana_datasource"
          }
        }
      }

      kubeStateMetrics = {
        enabled = true
      }

      nodeExporter = {
        enabled = true
      }
    })
  ]

  depends_on = [kubernetes_namespace.monitoring]
}

# -----------------------------------------------------------------------------
# Loki for Centralized Logging
# -----------------------------------------------------------------------------
resource "helm_release" "loki" {
  name       = "loki"
  repository = "https://grafana.github.io/helm-charts"
  chart      = "loki-stack"
  version    = var.loki_stack_version
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  values = [
    yamlencode({
      loki = {
        persistence = {
          enabled          = true
          storageClassName = var.storage_class
          size             = var.loki_storage_size
        }

        config = {
          schema_config = {
            configs = [
              {
                from         = "2024-01-01"
                store        = "boltdb-shipper"
                object_store = "s3"
                schema       = "v12"
                index = {
                  prefix = "loki_index_"
                  period = "24h"
                }
              }
            ]
          }

          storage_config = {
            aws = {
              s3               = "s3://${var.region}/${var.loki_s3_bucket}"
              s3forcepathstyle = true
            }
            boltdb_shipper = {
              active_index_directory = "/data/loki/boltdb-shipper-active"
              cache_location         = "/data/loki/boltdb-shipper-cache"
              shared_store           = "s3"
            }
          }
        }
      }

      promtail = {
        enabled = true
      }
    })
  ]

  depends_on = [kubernetes_namespace.monitoring]
}

# -----------------------------------------------------------------------------
# Multi-Region Dashboard ConfigMap
# -----------------------------------------------------------------------------
resource "kubernetes_config_map" "region_dashboard" {
  count = var.is_primary_region ? 1 : 0

  metadata {
    name      = "grafana-multi-region-dashboard"
    namespace = local.namespace
    labels = {
      grafana_dashboard = "1"
    }
  }

  data = {
    "multi-region-overview.json" = jsonencode({
      dashboard = {
        title = "VirtEngine Multi-Region Overview"
        uid   = "virtengine-multi-region"
        panels = [
          {
            title      = "Region Health"
            type       = "stat"
            datasource = "Prometheus"
            targets = [
              {
                expr = "up{job=~\"federation-.*\"}"
              }
            ]
            gridPos = {
              h = 4
              w = 24
              x = 0
              y = 0
            }
          },
          {
            title      = "Cross-Region Latency"
            type       = "timeseries"
            datasource = "Prometheus"
            targets = [
              {
                expr   = "histogram_quantile(0.99, rate(virtengine_cross_region_latency_seconds_bucket[5m]))"
                legend = "{{region}}"
              }
            ]
            gridPos = {
              h = 8
              w = 12
              x = 0
              y = 4
            }
          },
          {
            title      = "Database Replication Lag"
            type       = "timeseries"
            datasource = "Prometheus"
            targets = [
              {
                expr   = "cockroachdb_replication_lag_seconds"
                legend = "{{region}}"
              }
            ]
            gridPos = {
              h = 8
              w = 12
              x = 12
              y = 4
            }
          },
        ]
      }
    })
  }

  depends_on = [helm_release.prometheus_stack]
}

# -----------------------------------------------------------------------------
# Cross-Region Alert Rules
# -----------------------------------------------------------------------------
resource "kubernetes_manifest" "cross_region_alerts" {
  manifest = {
    apiVersion = "monitoring.coreos.com/v1"
    kind       = "PrometheusRule"
    metadata = {
      name      = "cross-region-alerts"
      namespace = local.namespace
      labels = {
        "app.kubernetes.io/name" = "virtengine-alerts"
      }
    }
    spec = {
      groups = [
        {
          name = "cross-region"
          rules = [
            {
              alert = "RegionDown"
              expr  = "up{job=~\"federation-.*\"} == 0"
              for   = "5m"
              labels = {
                severity = "critical"
              }
              annotations = {
                summary     = "Region {{ $labels.federated_region }} is down"
                description = "Federation target for region {{ $labels.federated_region }} has been unreachable for 5 minutes."
              }
            },
            {
              alert = "HighReplicationLag"
              expr  = "cockroachdb_replication_lag_seconds > 5"
              for   = "2m"
              labels = {
                severity = "warning"
              }
              annotations = {
                summary     = "High database replication lag in {{ $labels.region }}"
                description = "CockroachDB replication lag is {{ $value }}s in region {{ $labels.region }}."
              }
            },
            {
              alert = "CrossRegionLatencyHigh"
              expr  = "histogram_quantile(0.99, rate(virtengine_cross_region_latency_seconds_bucket[5m])) > 0.5"
              for   = "10m"
              labels = {
                severity = "warning"
              }
              annotations = {
                summary     = "High cross-region latency to {{ $labels.region }}"
                description = "P99 cross-region latency to {{ $labels.region }} is {{ $value }}s."
              }
            },
          ]
        }
      ]
    }
  }

  depends_on = [helm_release.prometheus_stack]
}
