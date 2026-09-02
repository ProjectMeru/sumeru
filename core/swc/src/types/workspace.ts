/** Workspace JSON from GET /web/swc/workspace */

export interface SwcViewArch {
  type: string;
  model: string;
  title?: string;
  hasChatter?: boolean;
  fields: SwcArchField[];
  header?: SwcArchHeader;
  footer?: SwcArchFooter;
  sheet?: SwcArchSheet;
  formMeta?: SwcFormMeta;
  kanban?: SwcKanbanMeta;
  pivot?: SwcPivotMeta;
  search?: SwcSearchMeta;
  graph?: SwcGraphMeta;
  hierarchy?: SwcHierarchyMeta;
  activity?: SwcActivityMeta;
  calendar?: SwcCalendarMeta;
  gantt?: SwcGanttMeta;
  map?: SwcMapMeta;
  cohort?: SwcCohortMeta;
  report?: SwcReportMeta;
}

export interface SwcFormMeta {
  hasImageField?: boolean;
}

export interface SwcArchFooter {
  buttons: SwcArchButton[];
}

export interface SwcArchSeparator {
  string?: string;
}

export interface SwcArchLabel {
  for?: string;
  string?: string;
}

export interface SwcArchListSubview {
  editable?: string;
  fields: SwcArchField[];
}

export interface SwcArchField {
  name: string;
  string?: string;
  type?: string;
  widget?: string;
  placeholder?: string;
  readonly?: boolean;
  required?: boolean;
  invisible?: boolean;
  readonly_expr?: string;
  required_expr?: string;
  invisible_expr?: string;
  default?: unknown;
  pivotType?: string;
  relation?: string;
  selection?: string[][];
  options?: Record<string, string>;
  subview?: SwcArchListSubview;
}

export interface SwcArchButton {
  name: string;
  string: string;
  type: "object" | "action" | string;
  class?: string;
  invisible?: boolean;
  invisible_expr?: string;
  confirm?: string;
}

export interface SwcArchHeader {
  buttons: SwcArchButton[];
  fields: SwcArchField[];
}

export interface SwcArchSheet {
  divs?: SwcArchDiv[];
  groups: SwcArchGroup[];
  notebook?: SwcArchNotebook[];
  fields: SwcArchField[];
  separators?: SwcArchSeparator[];
  labels?: SwcArchLabel[];
}

export interface SwcArchDiv {
  class?: string;
  fields?: SwcArchField[];
  buttons?: SwcArchButton[];
  h1Fields?: SwcArchField[];
  divs?: SwcArchDiv[];
}

export interface SwcArchGroup {
  string?: string;
  col?: number;
  colspan?: number;
  fields: SwcArchField[];
  groups?: SwcArchGroup[];
  separators?: SwcArchSeparator[];
  labels?: SwcArchLabel[];
}

export interface SwcArchNotebook {
  pages: SwcArchPage[];
}

export interface SwcArchPage {
  title: string;
  groups: SwcArchGroup[];
  fields: SwcArchField[];
  separators?: SwcArchSeparator[];
  labels?: SwcArchLabel[];
}

export interface SwcKanbanMeta {
  groupField: string;
  draggable: boolean;
  quickCreate?: boolean;
  columns: SwcKanbanColumn[];
}

export interface SwcKanbanColumn {
  value: number;
  label: string;
  sequence: number;
  color: number;
  rottingDays?: number;
  fold: boolean;
  records: Record<string, unknown>[];
}

export interface SwcPivotMeta {
  rowLabels: string[];
  colLabels: string[];
  values: Record<string, Record<string, number>>;
  measureLabel: string;
}

export interface SwcSearchFilter {
  name: string;
  string: string;
  domain?: string;
  groupBy?: string;
}

export interface SwcSearchMeta {
  filters?: SwcSearchFilter[];
  searchFields?: SwcArchField[];
  filterFields?: SwcArchField[];
  groupByFields?: SwcArchField[];
}

export interface SwcSavedSearch {
  id: number;
  name: string;
  search?: string;
  filter?: string;
  domain?: string;
  groupBy?: string;
  isDefault?: boolean;
  isShared?: boolean;
}

export interface SwcListSection {
  label: string;
  value: string;
  count: number;
  records: Record<string, unknown>[];
}

export interface SwcHierarchyMeta {
  parentField?: string;
}

export interface SwcActivityMeta {
  resModel?: string;
  resId?: string;
}

export interface SwcGraphMeta {
  chart?: string;
}

export interface SwcCalendarMeta {
  dateStart?: string;
  dateStop?: string;
}

export interface SwcGanttMeta {
  dateStart?: string;
  dateStop?: string;
}

export interface SwcMapMeta {
  latitude?: string;
  longitude?: string;
}

export interface SwcCohortMeta {
  dateStart?: string;
  interval?: string;
  measure?: string;
}

export interface SwcReportMeta {
  download: boolean;
  upload: boolean;
  formats?: string;
  pdfSizes: string;
  bulkModes: string;
}

export interface SwcViewTab {
  mode: string;
  label: string;
  href: string;
  active: boolean;
}

export interface SwcBreadcrumb {
  label: string;
  href?: string;
}

export interface SwcWorkspacePayload {
  actionId: number;
  menuId: string;
  viewType: string;
  model: string;
  recordId: number;
  formEdit: boolean;
  csrfToken: string;
  arch: SwcViewArch;
  record?: Record<string, unknown>;
  records?: Record<string, unknown>[];
  viewTabs: SwcViewTab[];
  breadcrumbs: SwcBreadcrumb[];
  listSearch?: string;
  listSearchUrl?: string;
  listTotal?: number;
  listSort?: string;
  listOffset?: number;
  listFilter?: string;
  listDomain?: string;
  listGroupBy?: string;
  listSections?: SwcListSection[];
  favorites?: SwcSavedSearch[];
  formBaseQuery?: string;
  defaults?: Record<string, unknown>;
}
