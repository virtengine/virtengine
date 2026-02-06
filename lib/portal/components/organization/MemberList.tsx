/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * MemberList — Displays organization members with role management.
 */

import { createElement } from "react";
import type {
  OrganizationMember,
  OrganizationRole,
} from "../../types/organization";
import { ROLE_LABELS } from "../../types/organization";

export interface MemberListProps {
  members: OrganizationMember[];
  isAdmin: boolean;
  currentUserAddress?: string;
  isRemoving?: boolean;
  onRemove: (address: string) => void;
  onUpdateRole?: (address: string, role: OrganizationRole) => void;
}

export function MemberList({
  members,
  isAdmin,
  currentUserAddress,
  isRemoving,
  onRemove,
}: MemberListProps): JSX.Element {
  if (members.length === 0) {
    return createElement(
      "div",
      { className: "py-8 text-center text-muted-foreground" },
      "No members found",
    );
  }

  return createElement(
    "div",
    { className: "space-y-2" },
    ...members.map((member) =>
      createElement(
        "div",
        {
          key: member.address,
          className: "flex items-center justify-between rounded-lg border p-3",
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
                className:
                  "text-sm text-destructive hover:underline disabled:opacity-50",
                onClick: () => onRemove(member.address),
                disabled: isRemoving,
              },
              "Remove",
            )
          : member.address === currentUserAddress
            ? createElement(
                "span",
                { className: "text-xs text-muted-foreground" },
                "You",
              )
            : null,
      ),
    ),
  );
}
