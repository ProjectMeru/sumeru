/** Coerce RPC/ORM boolean-ish values to a JS boolean. */
export function booleanFromUnknown(value: unknown): boolean {
  return value === true || value === 1 || value === "1" || value === "true";
}

/** Display string for a field value; empty when missing. */
export function stringFromUnknown(value: unknown): string {
  if (value == null || value === false) return "";
  return String(value);
}

/** Many2one-style label: `field_name` then `#id`. */
export function recordDisplayName(
  record: { get: (fieldName: string) => unknown },
  fieldName: string,
): string {
  const named = record.get(`${fieldName}_name`);
  if (named != null && named !== "") return String(named);
  const raw = record.get(fieldName);
  if (raw == null || raw === "") return "";
  return `#${raw}`;
}
