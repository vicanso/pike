import request from 'axios';
import _ from 'lodash';
import dayjs from 'dayjs';

import {
  STATS,
  DIRECTORS,
} from '../../urls';

import {
  PIKE_STATS,
  PIKE_DIRECTORS,
} from '../mutation-types';

const state = {
  stats: null,
  directors: null,
};

const mutations = {
  [PIKE_STATS](state, data) {
    data.startedAt = dayjs(data.startedAt).format('YYYY-MM-DD HH:mm');
    state.stats = data;
  },
  [PIKE_DIRECTORS](state, data) {
    _.forEach(data.directors, (item) => {
      const backends = [];
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

export const actions = {
  getStats,
  getDirectors,
};

export default {
  state,
  mutations,
};