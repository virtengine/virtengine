import { describe, expect, it } from "@jest/globals";

import {
  type FinancialCase,
  FinancialCaseStatus,
  FinancialClaimType,
  FinancialSubjectType,
  type MsgOpenFinancialCase,
  type QueryFinancialCaseRequest,
} from "../index.ts";

describe("canonical financial-case public exports", () => {
  it("exports aggregate, query, message, and enum contracts from the package root", () => {
    const message: MsgOpenFinancialCase = {
      sender: "virtengine1sender",
      subject: {
        type: FinancialSubjectType.FINANCIAL_SUBJECT_TYPE_ORDER,
        primaryId: "order-1",
        orderId: "order-1",
        invoiceId: "",
        usageId: "",
        hpcJobId: "",
        settlementId: "",
        escrowId: "",
        reservationId: "",
        leaseId: "",
      },
      claimType: FinancialClaimType.FINANCIAL_CLAIM_TYPE_BILLING,
      respondent: "virtengine1respondent",
      evidenceHash: new Uint8Array(32),
      encryptedReference: "",
      idempotencyKey: new Uint8Array([1]),
      sourceModule: "settlement",
      sourceReference: "order-1",
    };
    const request: QueryFinancialCaseRequest = { caseId: "financial-case/1" };
    const financialCase = { status: FinancialCaseStatus.FINANCIAL_CASE_STATUS_OPEN } as FinancialCase;

    expect(message.subject?.primaryId).toBe("order-1");
    expect(request.caseId).toBe("financial-case/1");
    expect(financialCase.status).toBe(FinancialCaseStatus.FINANCIAL_CASE_STATUS_OPEN);
  });
});
