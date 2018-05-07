import request from 'axios';
import _ from 'lodash';
import dayjs from 'dayjs';

import {
  STATS,
  DIRECTORS,
  CACHEDS,
} from '../../urls';

import {
  PIKE_STATS,
  PIKE_DIRECTORS,
  PIKE_CACHED,
  PIKE_CACHED_CLEAR,
} from '../mutation-types';

const state = {
  stats: null,
  directors: null,
  performances: null,
  cacheds: null,
};
const minute = 60;
const hour = 60 * minute;
const day = 24 * hour;

function getExpiredDesc(seconds) {
  if (seconds >= day) {
    return _.ceil(seconds / day, 2) + ' D';
  }
  if (seconds >= hour) {
    return _.ceil(seconds / hour, 2) + ' h';
  }
  if (seconds >= minute) {
    return _.ceil(seconds / minute, 2) + ' m';
  }
  return `${seconds} s`;
}

let requestCount = 0;
const performanceMaxCount = 60;
const mutations = {
  [PIKE_STATS](state, data) {
    data.startedAt = dayjs(data.startedAt).format('YYYY-MM-DD HH:mm');
    const performances = (state.performances || []).slice(0);
    const performance = _.omit(data, [
      'startedAt',
      'requestCount',
      'version',
    ]);
    performance.createdAt = Date.now();
    if (requestCount === 0 ) {
      performance.requestCount = 0;
    } else {
      performance.requestCount = data.requestCount - requestCount;
    }
    requestCount = data.requestCount;
    performances.push(performance);
    if (performances.length > performanceMaxCount) {
      performances.shift();
    }
    state.performances = performances;
    state.stats = data;
  },
  [PIKE_DIRECTORS](state, data) {
    _.forEach(data.directors, (item) => {
      const backends = [];
      // 生成backend列表
      _.forEach(item.backends, (backend) => {
        let status = 'sick';
        if (_.includes(item.availableBackends, backend)) {
          status = 'healthy';
        }
        backends.push({
          backend,
          status,
        });
      });
      item.backends = backends;
    })
    state.directors = data.directors;
  },
  [PIKE_CACHED](state, data) {
    const items = _.sortBy(data.cacheds, item => item.key);
    const now = Math.floor(Date.now() / 1000);
    _.forEach(items, (item) => {
      const {
        ttl,
        createdAt,
      } = item;
      item.createdAt = dayjs(createdAt * 1000).format('YYYY-MM-DD HH:mm:ss');
      const expiredSeconds = createdAt + ttl - now;
      item.expiredSeconds = expiredSeconds;
      item.expiredDesc = getExpiredDesc(expiredSeconds);
    });
    state.cacheds = items;
  },
  [PIKE_CACHED_CLEAR](state, key) {
    const cacheds = _.filter(state.cacheds, item => item.key != key);
    state.cacheds = cacheds;
  },
};

// 获取系统性能统计相关信息
async function getStats({commit}) {
  const res = await request.get(STATS);
  commit(PIKE_STATS, res.data);
}

// 获取directors相关信息
async function getDirectors({commit}) {
  const res = await request.get(DIRECTORS);
  commit(PIKE_DIRECTORS, res.data);
}

// 获取已缓存的接口列表
async function getCached({commit}) {
  const res = await request.get(CACHEDS);
  commit(PIKE_CACHED, res.data);
}

async function clearCached({commit}, key) {
  if (!window.btoa) {
    throw new Error('the browser is support btoa function, please upgrade')
  }
  const url = `${CACHEDS}/${window.btoa(key)}`
  await request.delete(url)
  commit(PIKE_CACHED_CLEAR, key);
}

export const actions = {
  getStats,
  getDirectors,
  getCached,
  clearCached,
};

export default {
  state,
  mutations,
};