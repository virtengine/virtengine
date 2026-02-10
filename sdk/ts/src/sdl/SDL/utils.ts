export function castArray<T>(value: T | T[] | undefined | null): T[] {
  if (value === undefined || value === null) return [];
  return Array.isArray(value) ? value : [value];
}

export function stringToBoolean(str: string | boolean): boolean {
  if (typeof str === "boolean") {
    return str;
  }

  switch (str.toLowerCase()) {
    case "false":
    case "no":
    case "0":
    case "":
      return false;
    default:
      return true;
  }
}
