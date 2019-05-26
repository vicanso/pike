import { getURLPrefix } from "./util";

const prefix = getURLPrefix();

export const UPSTREAMS = `${prefix}/upstreams`;
export const CACHES = `${prefix}/caches`;
export const STATS = `${prefix}/stats`;
export const CONFIGS = `${prefix}/configs`;
