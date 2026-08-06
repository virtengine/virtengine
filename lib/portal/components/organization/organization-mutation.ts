import type {
  CreateOrganizationRequest,
  InviteMemberRequest,
  Organization,
  OrganizationMember,
  OrganizationRole,
} from "../../types/organization";

export type OrganizationMutationAction =
  | "create"
  | "invite"
  | "remove"
  | "update_role"
  | "leave";

interface OrganizationMutationRequestBase {
  chainId: string;
  accountAddress: string;
  action: OrganizationMutationAction;
}

export type OrganizationMutationRequest =
  | (OrganizationMutationRequestBase & {
      action: "create";
      organization: CreateOrganizationRequest;
    })
  | (OrganizationMutationRequestBase & {
      action: "invite";
      organizationId: string;
      invitation: InviteMemberRequest;
    })
  | (OrganizationMutationRequestBase & {
      action: "remove";
      organizationId: string;
      memberAddress: string;
    })
  | (OrganizationMutationRequestBase & {
      action: "update_role";
      organizationId: string;
      memberAddress: string;
      role: OrganizationRole;
    })
  | (OrganizationMutationRequestBase & {
      action: "leave";
      organizationId: string;
    });

export interface OrganizationMutationContext {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}

export interface OrganizationMutationAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  mutateOrganization(
    request: OrganizationMutationRequest,
    context: OrganizationMutationContext,
  ): Promise<unknown>;
}

interface CommittedOrganizationMutationBase {
  status: "committed";
  code: 0;
  txHash: string;
  blockHeight: number;
  operationId: string;
  requestDigest: string;
  idempotencyKey: string;
  request: OrganizationMutationRequest;
}

export type CommittedOrganizationMutation =
  | (CommittedOrganizationMutationBase & {
      action: "create";
      organization: Organization;
    })
  | (CommittedOrganizationMutationBase & {
      action: "invite" | "remove" | "update_role";
      members: readonly OrganizationMember[];
    })
  | (CommittedOrganizationMutationBase & {
      action: "leave";
      organizationId: string;
    });

export class OrganizationMutationError extends Error {
  constructor(
    readonly code:
      | "feature_unavailable"
      | "invalid_request"
      | "invalid_committed_result"
      | "request_changed"
      | "submission_cancelled"
      | "submission_in_progress",
  ) {
    super(code);
    this.name = "OrganizationMutationError";
  }
}

const canonicalValue = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(canonicalValue);
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>)
        .filter(([, item]) => item !== undefined)
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, item]) => [key, canonicalValue(item)]),
    );
  }
  return value;
};
const canonical = (value: unknown): string =>
  JSON.stringify(canonicalValue(value));
const text = (value: unknown): string | null =>
  typeof value === "string" && value.trim() ? value.trim() : null;
const isRole = (value: unknown): value is OrganizationRole =>
  value === "admin" || value === "member" || value === "viewer";

const cloneFrozen = (value: unknown): unknown => {
  if (Array.isArray(value)) return Object.freeze(value.map(cloneFrozen));
  if (value && typeof value === "object") {
    if (value instanceof Date) return Object.freeze(new Date(value.getTime()));
    return Object.freeze(
      Object.fromEntries(
        Object.entries(value as Record<string, unknown>).map(([key, item]) => [
          key,
          cloneFrozen(item),
        ]),
      ),
    );
  }
  if (
    value === undefined ||
    value === null ||
    typeof value === "string" ||
    typeof value === "boolean" ||
    (typeof value === "number" && Number.isFinite(value))
  ) {
    return value;
  }
  throw new OrganizationMutationError("invalid_request");
};

const validCreateRequest = (value: CreateOrganizationRequest): boolean =>
  !!text(value.name) &&
  (value.description === undefined || typeof value.description === "string") &&
  (value.initialMembers === undefined ||
    (Array.isArray(value.initialMembers) &&
      value.initialMembers.every(
        (member) => !!text(member.address) && isRole(member.role),
      )));

export function buildOrganizationMutationRequest(
  action: OrganizationMutationAction,
  binding: { chainId: string; accountAddress: string },
  input:
    | CreateOrganizationRequest
    | { organizationId: string; invitation: InviteMemberRequest }
    | { organizationId: string; memberAddress: string }
    | {
        organizationId: string;
        memberAddress: string;
        role: OrganizationRole;
      }
    | { organizationId: string },
): OrganizationMutationRequest {
  if (!text(binding.chainId) || !text(binding.accountAddress)) {
    throw new OrganizationMutationError("invalid_request");
  }
  let request: OrganizationMutationRequest;
  const normalizedBinding = {
    chainId: binding.chainId.trim(),
    accountAddress: binding.accountAddress.trim(),
  };
  if (action === "create") {
    const organization = input as CreateOrganizationRequest;
    if (!validCreateRequest(organization)) {
      throw new OrganizationMutationError("invalid_request");
    }
    request = {
      ...normalizedBinding,
      action,
      organization: {
        name: organization.name.trim(),
        description: organization.description,
        initialMembers: organization.initialMembers?.map((member) => ({
          address: member.address.trim(),
          role: member.role,
        })),
      },
    };
  } else {
    const organizationId = text(
      (input as { organizationId?: unknown }).organizationId,
    );
    if (!organizationId) throw new OrganizationMutationError("invalid_request");
    if (action === "invite") {
      const invitation = (input as { invitation?: InviteMemberRequest })
        .invitation;
      if (
        !invitation ||
        !text(invitation.address) ||
        !isRole(invitation.role)
      ) {
        throw new OrganizationMutationError("invalid_request");
      }
      request = {
        ...normalizedBinding,
        action,
        organizationId,
        invitation: {
          address: invitation.address.trim(),
          role: invitation.role,
        },
      };
    } else if (action === "remove") {
      const memberAddress = text(
        (input as { memberAddress?: unknown }).memberAddress,
      );
      if (!memberAddress)
        throw new OrganizationMutationError("invalid_request");
      request = { ...normalizedBinding, action, organizationId, memberAddress };
    } else if (action === "update_role") {
      const update = input as {
        memberAddress?: unknown;
        role?: unknown;
      };
      const memberAddress = text(update.memberAddress);
      if (!memberAddress || !isRole(update.role)) {
        throw new OrganizationMutationError("invalid_request");
      }
      request = {
        ...normalizedBinding,
        action,
        organizationId,
        memberAddress,
        role: update.role,
      };
    } else if (action === "leave") {
      request = { ...normalizedBinding, action, organizationId };
    } else {
      throw new OrganizationMutationError("invalid_request");
    }
  }
  return cloneFrozen(request) as OrganizationMutationRequest;
}

export async function digestOrganizationMutationRequest(
  request: OrganizationMutationRequest,
): Promise<string> {
  const bytes = new TextEncoder().encode(canonical(request));
  const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
}

const materializeDate = (value: unknown): Date => {
  if (
    !(value instanceof Date) &&
    (typeof value !== "string" || !value.trim())
  ) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  const date =
    value instanceof Date ? new Date(value.getTime()) : new Date(value);
  if (!Number.isFinite(date.getTime())) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  if (typeof value === "string") {
    const match = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{3}))?Z$/.exec(
      value,
    );
    const canonicalInput = match ? `${match[1]}.${match[2] ?? "000"}Z` : null;
    if (!canonicalInput || date.toISOString() !== canonicalInput) {
      throw new OrganizationMutationError("invalid_committed_result");
    }
  }
  return Object.freeze(date);
};

const materializeOrganization = (value: unknown): Organization => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  const source = value as Partial<Organization>;
  const metadata = source.metadata as unknown as
    | Record<string, unknown>
    | undefined;
  if (
    !text(source.id) ||
    !text(source.name) ||
    !text(source.admin) ||
    !text(source.totalWeight) ||
    !metadata ||
    !text(metadata.name) ||
    (source.description !== undefined &&
      typeof source.description !== "string") ||
    (metadata.description !== undefined &&
      typeof metadata.description !== "string") ||
    (metadata.website !== undefined && typeof metadata.website !== "string") ||
    (metadata.logo !== undefined && typeof metadata.logo !== "string")
  ) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  return cloneFrozen({
    id: source.id!.trim(),
    name: source.name!.trim(),
    description: source.description,
    admin: source.admin!.trim(),
    totalWeight: source.totalWeight!.trim(),
    createdAt: materializeDate(source.createdAt),
    metadata,
  }) as Organization;
};

const materializeMember = (value: unknown): OrganizationMember => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  const source = value as Partial<OrganizationMember>;
  const metadata = source.metadata as unknown as
    | Record<string, unknown>
    | undefined;
  if (
    !text(source.address) ||
    !text(source.weight) ||
    !isRole(source.role) ||
    (metadata !== undefined &&
      (typeof metadata !== "object" ||
        metadata === null ||
        Array.isArray(metadata) ||
        (metadata.name !== undefined && typeof metadata.name !== "string") ||
        (metadata.email !== undefined && typeof metadata.email !== "string")))
  ) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  return cloneFrozen({
    address: source.address!.trim(),
    weight: source.weight!.trim(),
    role: source.role,
    addedAt: materializeDate(source.addedAt),
    metadata,
  }) as OrganizationMember;
};

export function validateCommittedOrganizationMutation(
  value: unknown,
  request: OrganizationMutationRequest,
  requestDigest: string,
): CommittedOrganizationMutation {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  const source = value as Record<string, unknown>;
  if (
    source.status !== "committed" ||
    source.code !== 0 ||
    !text(source.txHash) ||
    !Number.isSafeInteger(source.blockHeight) ||
    (source.blockHeight as number) <= 0 ||
    !text(source.operationId) ||
    source.requestDigest !== requestDigest ||
    source.idempotencyKey !== requestDigest ||
    canonical(source.request) !== canonical(request) ||
    source.action !== request.action
  ) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  const receipt = {
    status: "committed" as const,
    code: 0 as const,
    txHash: (source.txHash as string).trim(),
    blockHeight: source.blockHeight as number,
    operationId: (source.operationId as string).trim(),
    requestDigest,
    idempotencyKey: requestDigest,
    request,
  };
  if (request.action === "create") {
    return Object.freeze({
      ...receipt,
      action: request.action,
      organization: materializeOrganization(source.organization),
    });
  }
  if (
    request.action === "invite" ||
    request.action === "remove" ||
    request.action === "update_role"
  ) {
    if (!Array.isArray(source.members)) {
      throw new OrganizationMutationError("invalid_committed_result");
    }
    return Object.freeze({
      ...receipt,
      action: request.action,
      members: Object.freeze(source.members.map(materializeMember)),
    });
  }
  if (source.organizationId !== request.organizationId) {
    throw new OrganizationMutationError("invalid_committed_result");
  }
  return Object.freeze({
    ...receipt,
    action: request.action,
    organizationId: request.organizationId,
  });
}

export function requireOrganizationMutationAdapter(
  adapter: OrganizationMutationAdapter | undefined,
  binding: { chainId: string; accountAddress: string },
): OrganizationMutationAdapter {
  if (
    !adapter ||
    adapter.chainId !== binding.chainId ||
    adapter.accountAddress !== binding.accountAddress
  ) {
    throw new OrganizationMutationError("feature_unavailable");
  }
  return adapter;
}
