import { QueryAccountsRequest, QueryAccountsResponse, QueryInvoiceLedgerRequest, QueryInvoiceLedgerResponse, QueryInvoiceRequest, QueryInvoiceResponse, QueryInvoicesByCustomerRequest, QueryInvoicesByCustomerResponse, QueryInvoicesByProviderRequest, QueryInvoicesByProviderResponse, QueryPaymentsRequest, QueryPaymentsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.escrow.v1.Query",
  methods: {
    accounts: {
      name: "Accounts",
      httpPath: "/virtengine/escrow/v1/types/accounts",
      input: QueryAccountsRequest,
      output: QueryAccountsResponse,
      get parent() { return Query; },
    },
    payments: {
      name: "Payments",
      httpPath: "/virtengine/escrow/v1/types/payments",
      input: QueryPaymentsRequest,
      output: QueryPaymentsResponse,
      get parent() { return Query; },
    },
    invoice: {
      name: "Invoice",
      httpPath: "/virtengine/escrow/v1/billing/invoices/{invoice_id}",
      input: QueryInvoiceRequest,
      output: QueryInvoiceResponse,
      get parent() { return Query; },
    },
    invoicesByProvider: {
      name: "InvoicesByProvider",
      httpPath: "/virtengine/escrow/v1/billing/providers/{provider}/invoices",
      input: QueryInvoicesByProviderRequest,
      output: QueryInvoicesByProviderResponse,
      get parent() { return Query; },
    },
    invoicesByCustomer: {
      name: "InvoicesByCustomer",
      httpPath: "/virtengine/escrow/v1/billing/customers/{customer}/invoices",
      input: QueryInvoicesByCustomerRequest,
      output: QueryInvoicesByCustomerResponse,
      get parent() { return Query; },
    },
    invoiceLedger: {
      name: "InvoiceLedger",
      httpPath: "/virtengine/escrow/v1/billing/invoices/{invoice_id}/ledger",
      input: QueryInvoiceLedgerRequest,
      output: QueryInvoiceLedgerResponse,
      get parent() { return Query; },
    },
  },
} as const;
