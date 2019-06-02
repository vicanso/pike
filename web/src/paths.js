import { getURLPrefix } from "./util";

const prefix = getURLPrefix();

export const DIRECTOR_PATH = `${prefix}/`;
export const CACHES_PATH = `${prefix}/caches`;
export const PERFORMANCE_PATH = `${prefix}/performance`;
export const CONFIG_PATH = `${prefix}/config`;
export const ADD_UPSTREAM_PATH = `${prefix}/add-upstream`;
