export type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object ? DeepReadonly<T[P]> : T[P];
};

export type PickByPath<T, Path extends string> = Path extends `${infer First}.${infer Rest}`
  ? First extends keyof T
    ? { [K in First]: PickByPath<T[First], Rest> }
    : never
  : Path extends keyof T
    ? { [K in Path]: T[Path] }
    : never;
