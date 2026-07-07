/** RFC3339 / ISO 时间格式化为本地可读字符串 */
export function formatDateTime(value) {
  if (!value) return '-';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return String(value);
  return d.toLocaleString('zh-CN', { hour12: false });
}

/** 从 API 错误对象提取可读消息 */
export function getApiErrorMessage(error, fallback = '操作失败') {
  return error?.response?.data?.message || error?.message || fallback;
}
