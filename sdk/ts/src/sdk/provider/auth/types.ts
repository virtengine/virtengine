export interface JwtAuthentication {
  type: "jwt";
  token: string;
}

export interface mTLSAuthentication {
  type: "mtls";
  cert: string;
  key: string;
}
