declare var __DEV__: boolean;

declare interface Window {
  __DEV__?: boolean;
}

declare namespace NodeJS {
  interface Global {
    __DEV__?: boolean;
  }
}
