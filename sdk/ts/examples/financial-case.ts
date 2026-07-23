// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import { sha256 } from "@cosmjs/crypto";
import { toUtf8 } from "@cosmjs/encoding";

import type { MsgOpenFinancialCase } from "../src/generated/protos/virtengine/settlement/v1/tx.js";
import { FinancialClaimType, FinancialSubjectType } from "../src/generated/protos/virtengine/settlement/v1/query.js";
import type {
  QueryFinancialCaseRequest,
  QueryFinancialCaseResponse,
} from "../src/generated/protos/virtengine/settlement/v1/query.js";

interface FinancialCaseQueryClient {
  financialCase(request: QueryFinancialCaseRequest): Promise<QueryFinancialCaseResponse>;
}

export function buildFinancialCaseMessage(
  sender: string,
  respondent: string,
  orderId: string,
  encryptedReference: string,
  idempotencyKey: Uint8Array,
): MsgOpenFinancialCase {
  return {
    sender,
    respondent,
    subject: {
      type: FinancialSubjectType.FINANCIAL_SUBJECT_TYPE_ORDER,
      primaryId: orderId,
      orderId,
      invoiceId: "",
      usageId: "",
      hpcJobId: "",
      settlementId: "",
      escrowId: "",
      reservationId: "",
      leaseId: "",
    },
    claimType: FinancialClaimType.FINANCIAL_CLAIM_TYPE_BILLING,
    evidenceHash: sha256(toUtf8(encryptedReference)),
    encryptedReference,
    idempotencyKey,
    sourceModule: "sdk",
    sourceReference: orderId,
  };
}

export async function queryAuthoritativeCase(query: FinancialCaseQueryClient, caseId: string) {
  return query.financialCase({ caseId });
}