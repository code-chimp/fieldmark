export function projectSlotForBaseUrl(baseURL?: string) {
  if ((baseURL ?? '').includes('4000')) {
    return 0;
  }
  if ((baseURL ?? '').includes('8000')) {
    return 1;
  }
  if ((baseURL ?? '').includes('3000')) {
    return 2;
  }
  return 0;
}
