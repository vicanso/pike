import request from 'axios';

import {urlPrefix} from './config';
import {getAdminToken} from './helpers/util';

request.interceptors.request.use(config => {
  if (!config.timeout) {
    config.timeout = 10 * 1000;
  }
  try {
    const token = getAdminToken();
    config.headers['X-Admin-Token'] = token;
  } catch (err) {
    // eslint-disable-next-line
    console.error(err);
  }
  config.url = `${urlPrefix}${config.url}`;
  return config;
});
