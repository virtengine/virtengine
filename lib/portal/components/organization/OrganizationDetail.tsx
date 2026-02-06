/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * OrganizationDetail — Detail view for an organization with tabs.
 */

import { createElement, useState } from "react";
import type {
  Organization,
  OrganizationMember,
  OrganizationRole,
  OrganizationBillingSummary,
} from "../../types/organization";
import { hasPermission, ROLE_LABELS } from "../../types/organization";

export interface OrganizationDetailProps {
  organization: Organization | null;
  members: OrganizationMember[];
  billing: OrganizationBillingSummary | null;
  isLoading: boolean;
  error: string | null;
  currentUserRole: OrganizationRole | null;
  currentUserAddress?: string;
  onInviteMember: () => void;
  onRemoveMember: (address: string) => void;
  onUpdateRole: (address: string, role: OrganizationRole) => void;
  onLeave: () => void;
  onBack: () => void;
}

type DetailTab = "members" | "billing" | "settings";

export function OrganizationDetail({
  organization,
  members,
  billing,
  isLoading,
  error,
  currentUserRole,
  currentUserAddress,
  onInviteMember,
  onRemoveMember,
  onUpdateRole,
  onLeave,
  onBack,
}: OrganizationDetailProps): JSX.Element {
  const [activeTab, setActiveTab] = useState<DetailTab>("members");

  if (isLoading) {
    return createElement(
      "div",
      { className: "space-y-4" },
      createElement("div", {
        className: "h-8 w-48 animate-pulse rounded bg-muted",
      }),
      createElement("div", {
        className: "h-4 w-64 animate-pulse rounded bg-muted",
      }),
      createElement("div", {
        className: "h-64 animate-pulse rounded-lg bg-muted",
      }),
    );
  }

  if (error) {
    return createElement(
      "div",
      {
        className:
          "rounded-lg border border-destructive/50 bg-destructive/10 p-4",
        role: "alert",
      },
      createElement("p", { className: "text-sm text-destructive" }, error),
    );
  }

  if (!organization) {
    return createElement(
      "div",
      { className: "py-16 text-center text-muted-foreground" },
      "Organization not found",
    );
  }

  const isAdmin = currentUserRole === "admin";
  const canManageMembers = currentUserRole
    ? hasPermission(currentUserRole, "manage_members")
    : false;

  const tabs: { id: DetailTab; label: string; show: boolean }[] = [
    { id: "members", label: `Members (${members.length})`, show: true },
    {
      id: "billing",
      label: "Billing",
      show: currentUserRole
        ? hasPermission(currentUserRole, "view_billing")
        : false,
    },
    { id: "settings", label: "Settings", show: isAdmin },
  ];

  return createElement(
    "div",
    { className: "space-y-6" },
    // Header
    createElement(
      "div",
      { className: "flex items-start justify-between" },
      createElement(
        "div",
        null,
        createElement(
          "button",
          {
            type: "button",
            className:
              "mb-2 text-sm text-muted-foreground hover:text-foreground",
            onClick: onBack,
          },
          "← Back to organizations",
        ),
        createElement(
          "div",
          { className: "flex items-center gap-3" },
          createElement(
            "div",
            {
              className:
                "flex h-12 w-12 items-center justify-center rounded-full bg-primary/10",
            },
            createElement(
              "span",
              { className: "text-xl font-semibold text-primary" },
              organization.name.charAt(0).toUpperCase(),
            ),
          ),
          createElement(
            "div",
            null,
            createElement(
              "h1",
              { className: "text-2xl font-bold" },
              organization.name,
            ),
            organization.description
              ? createElement(
                  "p",
                  { className: "text-sm text-muted-foreground" },
                  organization.description,
                )
              : null,
          ),
        ),
      ),
      createElement(
        "div",
        { className: "flex items-center gap-2" },
        currentUserRole
          ? createElement(
              "span",
              {
                className:
                  "rounded-full bg-secondary px-3 py-1 text-sm capitalize text-secondary-foreground",
              },
              ROLE_LABELS[currentUserRole],
            )
          : null,
        !isAdmin
          ? createElement(
              "button",
              {
                type: "button",
                className:
                  "rounded-lg border border-destructive px-3 py-1 text-sm text-destructive hover:bg-destructive/10",
                onClick: onLeave,
              },
              "Leave",
            )
          : null,
      ),
    ),
    // Tabs
    createElement(
      "div",
      { className: "flex gap-4 border-b border-border", role: "tablist" },
      ...tabs
        .filter((t) => t.show)
        .map((tab) =>
          createElement(
            "button",
            {
              key: tab.id,
              type: "button",
              role: "tab",
              "aria-selected": activeTab === tab.id,
              className:
                activeTab === tab.id
                  ? "border-b-2 border-primary px-4 py-2 text-sm font-medium text-primary"
                  : "px-4 py-2 text-sm text-muted-foreground hover:text-foreground",
              onClick: () => setActiveTab(tab.id),
            },
            tab.label,
          ),
        ),
    ),
    // Tab Content
    activeTab === "members"
      ? createElement(MembersTab, {
          members,
          isAdmin: canManageMembers,
          currentUserAddress,
          onInvite: onInviteMember,
          onRemove: onRemoveMember,
          onUpdateRole,
        })
      : null,
    activeTab === "billing" ? createElement(BillingTab, { billing }) : null,
    activeTab === "settings"
      ? createElement(SettingsTab, { organization })
      : null,
  );
}

// =============================================================================
// Members Tab
// =============================================================================

interface MembersTabProps {
  members: OrganizationMember[];
  isAdmin: boolean;
  currentUserAddress?: string;
  onInvite: () => void;
  onRemove: (address: string) => void;
  onUpdateRole: (address: string, role: OrganizationRole) => void;
}

function MembersTab({
  members,
  isAdmin,
  currentUserAddress,
  onInvite,
  onRemove,
}: MembersTabProps): JSX.Element {
  return createElement(
    "div",
    { className: "space-y-4" },
    isAdmin
      ? createElement(
          "div",
          { className: "flex justify-end" },
          createElement(
            "button",
            {
              type: "button",
              className:
                "rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90",
              onClick: onInvite,
            },
            "+ Invite Member",
          ),
        )
      : null,
    createElement(
      "div",
      { className: "space-y-2" },
      ...members.map((member) =>
        createElement(
          "div",
          {
            key: member.address,
            className:
              "flex items-center justify-between rounded-lg border p-3",
          },
          createElement(
            "div",
            { className: "flex items-center gap-3" },
            createElement(
              "div",
              {
                className:
                  "flex h-8 w-8 items-center justify-center rounded-full bg-muted",
              },
              createElement(
                "span",
                { className: "text-sm" },
                member.address.slice(0, 2).toUpperCase(),
              ),
            ),
            createElement(
              "div",
              null,
              createElement(
                "p",
                { className: "font-mono text-sm" },
                `${member.address.slice(0, 12)}...${member.address.slice(-6)}`,
              ),
              createElement(
                "p",
                { className: "text-xs capitalize text-muted-foreground" },
                ROLE_LABELS[member.role],
              ),
            ),
          ),
          isAdmin && member.address !== currentUserAddress
            ? createElement(
                "button",
                {
                  type: "button",
                  className: "text-sm text-destructive hover:underline",
                  onClick: () => onRemove(member.address),
                },
                "Remove",
              )
            : null,
        ),
      ),
    ),
  );
}

// =============================================================================
// Billing Tab
// =============================================================================

function BillingTab({
  billing,
}: {
  billing: OrganizationBillingSummary | null;
}): JSX.Element {
  if (!billing) {
    return createElement(
      "div",
      { className: "py-8 text-center text-muted-foreground" },
      "Billing data not available",
    );
  }

  return createElement(
    "div",
    { className: "space-y-6" },
    // Summary cards
    createElement(
      "div",
      { className: "grid gap-4 md:grid-cols-3" },
      createElement(BillingStatCard, {
        label: "Current Period",
        value: `$${billing.currentPeriodSpend.toFixed(2)}`,
      }),
      createElement(BillingStatCard, {
        label: "Previous Period",
        value: `$${billing.previousPeriodSpend.toFixed(2)}`,
      }),
      createElement(BillingStatCard, {
        label: "Total Spend",
        value: `$${billing.totalSpend.toFixed(2)}`,
      }),
    ),
    // By member breakdown
    billing.byMember.length > 0
      ? createElement(
          "div",
          { className: "space-y-2" },
          createElement(
            "h3",
            { className: "text-sm font-medium" },
            "Breakdown by Member",
          ),
          ...billing.byMember.map((entry) =>
            createElement(
              "div",
              {
                key: entry.address,
                className:
                  "flex items-center justify-between rounded-lg border p-3",
              },
              createElement(
                "span",
                { className: "font-mono text-sm" },
                `${entry.address.slice(0, 12)}...${entry.address.slice(-6)}`,
              ),
              createElement(
                "div",
                { className: "text-right" },
                createElement(
                  "p",
                  { className: "text-sm font-medium" },
                  `$${entry.amount.toFixed(2)}`,
                ),
                createElement(
                  "p",
                  { className: "text-xs text-muted-foreground" },
                  `${entry.percentage.toFixed(1)}%`,
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
          { className: "space-y-2" },
          createElement("h3", { className: "text-sm font-medium" }, "History"),
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
                  `${period.deployments} deployments`,
                ),
              ),
              createElement(
                "p",
                { className: "text-sm font-medium" },
                `$${period.amount.toFixed(2)}`,
              ),
            ),
          ),
        )
      : null,
  );
}

function BillingStatCard({
  label,
  value,
}: {
  label: string;
  value: string;
}): JSX.Element {
  return createElement(
    "div",
    { className: "rounded-lg border p-4" },
    createElement("p", { className: "text-sm text-muted-foreground" }, label),
    createElement("p", { className: "mt-1 text-2xl font-semibold" }, value),
  );
}

// =============================================================================
// Settings Tab
// =============================================================================

function SettingsTab({
  organization,
}: {
  organization: Organization;
}): JSX.Element {
  return createElement(
    "div",
    { className: "space-y-4" },
    createElement(
      "div",
      { className: "rounded-lg border p-4" },
      createElement(
        "h3",
        { className: "text-sm font-medium" },
        "Organization Info",
      ),
      createElement(
        "div",
        { className: "mt-2 space-y-2" },
        createElement(
          "div",
          null,
          createElement(
            "p",
            { className: "text-xs text-muted-foreground" },
            "Name",
          ),
          createElement("p", { className: "text-sm" }, organization.name),
        ),
        createElement(
          "div",
          null,
          createElement(
            "p",
            { className: "text-xs text-muted-foreground" },
            "Admin",
          ),
          createElement(
            "p",
            { className: "font-mono text-sm" },
            organization.admin,
          ),
        ),
        createElement(
          "div",
          null,
          createElement(
            "p",
            { className: "text-xs text-muted-foreground" },
            "Created",
          ),
          createElement(
            "p",
            { className: "text-sm" },
            organization.createdAt.toLocaleDateString(),
          ),
        ),
        organization.metadata.website
          ? createElement(
              "div",
              null,
              createElement(
                "p",
                { className: "text-xs text-muted-foreground" },
                "Website",
              ),
              createElement(
                "p",
                { className: "text-sm" },
                organization.metadata.website,
              ),
            )
          : null,
      ),
    ),
  );
}
