/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * InviteMemberDialog — Dialog for inviting a new member to an organization.
 */

import { createElement, useState } from "react";
import type {
  OrganizationRole,
  InviteMemberRequest,
} from "../../types/organization";
import { ROLE_LABELS, ROLE_DESCRIPTIONS } from "../../types/organization";

export interface InviteMemberDialogProps {
  open: boolean;
  isPending: boolean;
  onClose: () => void;
  onInvite: (request: InviteMemberRequest) => void;
}

const ROLES: OrganizationRole[] = ["admin", "member", "viewer"];

export function InviteMemberDialog({
  open,
  isPending,
  onClose,
  onInvite,
}: InviteMemberDialogProps): JSX.Element | null {
  const [address, setAddress] = useState("");
  const [role, setRole] = useState<OrganizationRole>("member");

  if (!open) return null;

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (!address.trim()) return;
    onInvite({ address: address.trim(), role });
    setAddress("");
    setRole("member");
  };

  return createElement(
    "div",
    {
      className:
        "fixed inset-0 z-50 flex items-center justify-center bg-black/50",
      onClick: (e: MouseEvent) => {
        if ((e.target as HTMLElement) === e.currentTarget) onClose();
      },
    },
    createElement(
      "div",
      {
        className:
          "w-full max-w-md rounded-lg border bg-background p-6 shadow-lg",
        role: "dialog",
        "aria-modal": true,
        "aria-labelledby": "invite-dialog-title",
      },
      createElement(
        "h2",
        { id: "invite-dialog-title", className: "text-lg font-semibold" },
        "Invite Member",
      ),
      createElement(
        "form",
        { className: "mt-4 space-y-4", onSubmit: handleSubmit },
        // Address input
        createElement(
          "div",
          null,
          createElement(
            "label",
            { className: "text-sm font-medium", htmlFor: "invite-address" },
            "Wallet Address",
          ),
          createElement("input", {
            id: "invite-address",
            type: "text",
            className:
              "mt-1 flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
            placeholder: "virtengine1...",
            value: address,
            onChange: (e: Event) =>
              setAddress((e.target as HTMLInputElement).value),
            required: true,
          }),
        ),
        // Role selection
        createElement(
          "div",
          null,
          createElement("label", { className: "text-sm font-medium" }, "Role"),
          createElement(
            "div",
            { className: "mt-1 space-y-2" },
            ...ROLES.map((r) =>
              createElement(
                "label",
                {
                  key: r,
                  className: `flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors ${
                    role === r
                      ? "border-primary bg-primary/5"
                      : "hover:bg-muted/50"
                  }`,
                },
                createElement("input", {
                  type: "radio",
                  name: "role",
                  value: r,
                  checked: role === r,
                  onChange: () => setRole(r),
                  className: "mt-0.5",
                }),
                createElement(
                  "div",
                  null,
                  createElement(
                    "p",
                    { className: "text-sm font-medium" },
                    ROLE_LABELS[r],
                  ),
                  createElement(
                    "p",
                    { className: "text-xs text-muted-foreground" },
                    ROLE_DESCRIPTIONS[r],
                  ),
                ),
              ),
            ),
          ),
        ),
        // Actions
        createElement(
          "div",
          { className: "flex justify-end gap-2 pt-2" },
          createElement(
            "button",
            {
              type: "button",
              className: "rounded-lg border px-4 py-2 text-sm hover:bg-muted",
              onClick: onClose,
            },
            "Cancel",
          ),
          createElement(
            "button",
            {
              type: "submit",
              className:
                "rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50",
              disabled: isPending || !address.trim(),
            },
            isPending ? "Inviting..." : "Invite",
          ),
        ),
      ),
    ),
  );
}
