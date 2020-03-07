let prefix = window.location.pathname;
// 如果pathname是/，则使用默认前缀/admin/
if (prefix === '/') {
  prefix = '/admin/';
}

export const CONFIGS = `${prefix}configs/:category`;
export const CONFIG = `${prefix}configs/:category/:name`;
