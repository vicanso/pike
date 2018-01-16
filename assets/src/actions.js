import moment from 'moment';
import 'Base64';

const statsUrl = '/pike/stats';
const directorsUrl = '/pike/directors';

export const state = {
  concurrency: 0,
  sys: 0,
  heapSys: 0,
  heapInuse: 0,
  startedAt: '',
  routine: 0,
  cacheCount: 0,
  fetching: 0,
  waiting: 0,
  cacheable: 0,
  hitForPass: 0,
  view: 'default',
  launchedAt: '',
  uptime: '',
  directors: null,
};

export function getStats() {
  return fetch(statsUrl).then((res) => {
    return res.json();
  });
}

export function getDirectors() {
  return fetch(directorsUrl).then((res) => {
    return res.json();
  }).then((data) => {
    const covert = (item, key) => {
      if (item[key]) {
        item[key] = item[key].map(window.atob);
      }
    }
    data.forEach((item) => {
      ['hosts', 'passes', 'prefixs'].forEach(key => covert(item, key));
    });
    return data;
  });
}

export const actions = {
  resetDirectors: () => state => {
    return {
      directors: null,
    };
  },
  setStats: data => state => data,
  setLaunchedAt: launchedAt => state => {
    return {
      launchedAt: moment(launchedAt).format('YYYY-MM-DD HH:mm:ss'),
      uptime: moment(launchedAt).fromNow(),
    };
  },
  setDirectors: data => state => {
    return {
      directors: data,
    };
  },
  changeView: view => state => {
    return {
      view,
    };
  },
}
