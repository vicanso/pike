import _ from 'lodash';
import jsonDiff from 'json-diff';

import {sha256} from './crypto';
import {app} from '../config';

// 获取出错信息
export function getErrorMessage(err) {
  let message = err;
  if (err && err.response) {
    const {data} = err.response;
    message = data.message;
  }
  if (_.isError(message)) {
    message = message.message;
  }
  if (err.code === 'ECONNABORTED') {
    message = '请求超时，请重新再试';
  }
  return message;
}

// 生成密码
export function genPassword(account, password) {
  const pwd = sha256(password);
  return sha256(`${account}-${pwd}-${app}`);
}

// 获取日期格式化字符串
export function getDate(str) {
  const date = new Date(str);
  const fill = v => {
    if (v >= 10) {
      return `${v}`;
    }
    return `0${v}`;
  };
  const month = fill(date.getMonth() + 1);
  const day = fill(date.getDate());
  const hours = fill(date.getHours());
  const mintues = fill(date.getMinutes());
  const seconds = fill(date.getSeconds());
  return `${date.getFullYear()}-${month}-${day} ${hours}:${mintues}:${seconds}`;
}

// 等待ttl时长
export function waitFor(ttl, startedAt) {
  let delay = ttl;
  if (startedAt) {
    delay = ttl - (Date.now() - startedAt);
  }
  return new Promise(resolve => {
    setTimeout(resolve, Math.max(0, delay));
  });
}

// 获取不一致的数据
export function diff(original, current, keys) {
  const changeKeys = jsonDiff.diff(original, current);
  const diffKeys = keys || _.keys(current);
  const result = {};
  _.forEach(diffKeys, key => {
    if (!changeKeys[key]) {
      return;
    }
    result[key] = current[key];
  });
  return result;
}

const token = 'adminToken';

export function saveAdminToken(value) {
  if (!window.localStorage) {
    throw new Error('the browser is not support local storage');
  }
  localStorage.setItem(token, value);
}

export function getAdminToken() {
  if (!window.localStorage) {
    throw new Error('the browser is not support local storage');
  }
  return localStorage.getItem(token);
}
