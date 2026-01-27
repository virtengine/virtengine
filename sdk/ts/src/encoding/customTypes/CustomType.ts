export interface CustomType<SourceType, TargetType> {
  typeName: string;
  shortName: string;
  encode(value: SourceType): TargetType;
  decode(value: TargetType): SourceType;
}
