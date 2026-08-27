/** Read the current value of an input/textarea/select from a DOM event. */
export function inputValueFromEvent(event: Event): string {
  const target = event.target;
  if (
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement
  ) {
    return target.value;
  }
  return "";
}

/** Checkbox checked state from a change event. */
export function checkboxCheckedFromEvent(event: Event): boolean {
  const target = event.target;
  return target instanceof HTMLInputElement ? target.checked : false;
}
