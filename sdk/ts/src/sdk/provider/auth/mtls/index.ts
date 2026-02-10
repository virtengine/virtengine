import { CertificateManager } from "./CertificateManager.ts";

export { CertificateManager, type CertificateInfo, type CertificatePem, type ValidityRangeOptions } from "./CertificateManager.ts";
export const certificateManager = new CertificateManager();
