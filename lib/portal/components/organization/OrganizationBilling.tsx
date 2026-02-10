/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * OrganizationBilling — Billing aggregation view for organizations.
 */

import { createElement } from "react";
import type { OrganizationBillingSummary } from "../../types/organization";

export interface OrganizationBillingProps {
  billing: OrganizationBillingSummary | null;
  isLoading: boolean;
}

export function OrganizationBilling({
  billing,
  isLoading,
}: OrganizationBillingProps): JSX.Element {
  if (isLoading) {
    return createElement(
      "div",
      { className: "space-y-4" },
      createElement("div", {
        className: "h-24 animate-pulse rounded-lg bg-muted",
      }),
      createElement("div", {
        className: "h-24 animate-pulse rounded-lg bg-muted",
      }),
      createElement("div", {
        className: "h-24 animate-pulse rounded-lg bg-muted",
      }),
    );
  }

  if (!billing) {
    return createElement(
      "div",
      { className: "py-8 text-center text-muted-foreground" },
      "No billing data available",
    );
  }

  const changeColor =
    billing.changePercent > 0
      ? "text-destructive"
      : billing.changePercent < 0
        ? "text-success"
        : "text-muted-foreground";

  return createElement(
    "div",
    { className: "space-y-6" },
    // Stats cards
    createElement(
      "div",
      { className: "grid gap-4 md:grid-cols-4" },
      createElement(StatCard, {
        label: "Current Period",
        value: `$${billing.currentPeriodSpend.toFixed(2)}`,
      }),
      createElement(StatCard, {
        label: "Previous Period",
        value: `$${billing.previousPeriodSpend.toFixed(2)}`,
      }),
      createElement(StatCard, {
        label: "Change",
        value: `${billing.changePercent > 0 ? "+" : ""}${billing.changePercent.toFixed(1)}%`,
        valueClassName: changeColor,
      }),
      createElement(StatCard, {
        label: "Total Spend",
        value: `$${billing.totalSpend.toFixed(2)}`,
      }),
    ),
    // Member breakdown
    billing.byMember.length > 0
      ? createElement(
          "div",
          null,
          createElement(
            "h3",
            { className: "mb-3 text-sm font-medium" },
            "By Member",
          ),
          createElement(
            "div",
            { className: "overflow-hidden rounded-lg border" },
            createElement(
              "table",
              { className: "w-full" },
              createElement(
                "thead",
                null,
                createElement(
                  "tr",
                  { className: "border-b bg-muted/50" },
                  createElement(
                    "th",
                    {
                      className:
                        "px-4 py-2 text-left text-xs font-medium text-muted-foreground",
                    },
                    "Member",
                  ),
                  createElement(
                    "th",
                    {
                      className:
                        "px-4 py-2 text-right text-xs font-medium text-muted-foreground",
                    },
                    "Amount",
                  ),
                  createElement(
                    "th",
                    {
                      className:
                        "px-4 py-2 text-right text-xs font-medium text-muted-foreground",
                    },
                    "Share",
                  ),
                ),
              ),
              createElement(
                "tbody",
                null,
                ...billing.byMember.map((entry) =>
                  createElement(
                    "tr",
                    { key: entry.address, className: "border-b last:border-0" },
                    createElement(
                      "td",
                      { className: "px-4 py-2 font-mono text-sm" },
                      `${entry.address.slice(0, 12)}...${entry.address.slice(-6)}`,
                    ),
                    createElement(
                      "td",
                      { className: "px-4 py-2 text-right text-sm" },
                      `$${entry.amount.toFixed(2)}`,
                    ),
                    createElement(
                      "td",
                      {
                        className:
                          "px-4 py-2 text-right text-sm text-muted-foreground",
                      },
                      `${entry.percentage.toFixed(1)}%`,
                    ),
                  ),
                ),
              ),
            ),
          ),
        )
      : null,
    // History
    billing.history.length > 0
      ? createElement(
          "div",
          null,
          createElement(
            "h3",
            { className: "mb-3 text-sm font-medium" },
            "Invoice History",
          ),
          createElement(
            "div",
            { className: "space-y-2" },
            ...billing.history.map((period) =>
              createElement(
                "div",
                {
                  key: period.period,
                  className:
                    "flex items-center justify-between rounded-lg border p-3",
                },
                createElement(
                  "div",
                  null,
                  createElement(
                    "p",
                    { className: "text-sm font-medium" },
                    period.period,
                  ),
                  createElement(
                    "p",
                    { className: "text-xs text-muted-foreground" },
                    `${period.deployments} deployment${period.deployments !== 1 ? "s" : ""}`,
                  ),
                ),
                createElement(
                  "p",
                  { className: "text-sm font-medium" },
                  `$${period.amount.toFixed(2)}`,
                ),
              ),
            ),
          ),
        )
      : null,
  );
}

function StatCard({
  label,
  value,
  valueClassName,
}: {
  label: string;
  value: string;
  valueClassName?: string;
}): JSX.Element {
  return createElement(
    "div",
    { className: "rounded-lg border p-4" },
    createElement("p", { className: "text-sm text-muted-foreground" }, label),
    createElement(
      "p",
      { className: `mt-1 text-2xl font-semibold ${valueClassName || ""}` },
      value,
    ),
  );
}
