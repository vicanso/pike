import moment from 'moment';
import 'Base64';

let prefixUrl = '..';
if (location.search.indexOf('dev') !== -1) {
  prefixUrl = '/pike';
}

const statsUrl = prefixUrl + '/stats';
const directorsUrl = prefixUrl + '/directors';
const blockIPsUrl = prefixUrl + '/block-ips';
const upstreamsUrl = prefixUrl + '/upstreams';
const cachedsUrl = prefixUrl + '/cacheds';

const maxPointCount = 30;
const performanceKeys = [
  'concurrency',
  'sys',
  'heapSys',
  'heapInuse',
  'routine',
  'cacheCount',
  'fetching',
  'waiting',
  'cacheable',
  'hitForPass',
  'requestCount',
  'lsm',
  'vLog',
];

const adminToken = localStorage.getItem('adminToken');
const defaultHeader = {
  'X-Admin-Token': adminToken,
};

export const state = {
  performance: null,
  view: 'default',
  launchedAt: '',
  uptime: '',
  directors: null,
  blockIPList: null,
  setCacheds: null,
};

export function getStats() {
  return fetch(statsUrl, {
    headers: defaultHeader,
  }).then((res) => {
    if (res.status >= 400) {
      throw res
    }
    return res.json();
  });
}

export function getDirectors() {
  let directors = null;
  return fetch(directorsUrl, {
    headers: defaultHeader,
  }).then((res) => {
    if (res.status >= 400) {
      throw res
    }
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
    directors = data;
    return fetch(upstreamsUrl, {
      headers: defaultHeader,
    })
  }).then((res) => {
    if (res.status >= 400) {
      throw res
    }
    return res.json();
  }).then((data) => {
    directors.forEach((item) => {
      data.forEach((tmp) => {
        if (tmp.name === item.name) {
          item.upstream = tmp;
        }
      });
    });
    return directors;
  });
}

export function getBlockIPs() {
  return fetch(blockIPsUrl, {
    headers: defaultHeader,
  }).then((res) => {
    if (res.status >= 400) {
      throw res
    }
    return res.json();
  });
}

export function getCacheds() {
  return fetch(cachedsUrl, {
    headers: defaultHeader,
  }).then((res) => {
    if (res.status >= 400) {
      throw res
    }
    return res.json();
  });
}

export function addBlockIP(ip) {
  return fetch(blockIPsUrl, {
    method: 'POST',
    headers: defaultHeader,
    body: JSON.stringify({
      ip,
    }),
  }).then((res) => {
    if (res.status >= 400) {
      throw res;
    }
  });
}

export function removeBlockIP(ip) {
  return fetch(blockIPsUrl, {
    method: 'DELETE',
    headers: defaultHeader,
    body: JSON.stringify({
      ip,
    }),
  }).then((res) => {
    if (res.status >= 400) {
      throw res;
    }
  });
}

export function removeCache(key) {
  return fetch(cachedsUrl, {
    method: 'DELETE',
    headers: defaultHeader,
    body: JSON.stringify({
      key,
    }),
  }).then((res) => {
    if (res.status >= 400) {
      throw res;
    }
  });
}

export const actions = {
  resetDirectors: () => () => {
    return {
      directors: null,
    };
  },
  setPerformance: data => state => {
    const result = {};
    const prevPerformance = state.performance || {};
    const now = moment().format('HH:mm');
    performanceKeys.forEach((key) => {
      const arr = (prevPerformance[key] || []).slice(0);
      const value = data[key];
      let prev = null;
      if (arr.length === maxPointCount) {
        prev = arr.shift()
      }
      if (key === 'requestCount') {
        let v = 0;
        let last = arr[arr.length - 1] || prev;
        if (last) {
          v = value - last.count;
        }
        arr.push({
          time: now,
          count: value,
          value: v,
        });
      } else {
        arr.push({
          time: now,
          value,
        });
      }

      result[key] = arr;
    });
    return {
      performance: result,
    };
  },
  setLaunchedAt: launchedAt => () => {
    return {
      launchedAt: moment(launchedAt).format('YYYY-MM-DD HH:mm:ss'),
      uptime: moment(launchedAt).fromNow(),
    };
  },
  setDirectors: data => () => {
    return {
      directors: data,
    };
  },
  setBlockIPList: data => () => {
    return {
      blockIPList: data,
    };
  },
  changeView: view => () => {
    return {
      view,
    };
  },
  setCacheds: data => () => {
    return {
      cacheds: data,
    };
  },
  resetCacheds: () => () => {
    return {
      cacheds: null,
    };
  },
}
