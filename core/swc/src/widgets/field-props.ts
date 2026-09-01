import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";

/** Props shared by every field widget in the registry. */
export interface FieldWidgetProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
  modelName?: string;
}
