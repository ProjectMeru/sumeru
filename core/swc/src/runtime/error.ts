export class SwcError extends Error {
  readonly code: string;
  readonly details?: unknown;

  constructor(message: string, code = "swc_error", details?: unknown) {
    super(message);
    this.name = "SwcError";
    this.code = code;
    this.details = details;
  }
}

export function isSwcError(err: unknown): err is SwcError {
  return err instanceof SwcError;
}
