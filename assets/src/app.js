import { h, app } from 'hyperapp';
import 'whatwg-fetch';

import './global.sss'
import './app.sss'

import * as chart from './chart';

const state = {
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
};
const statsUrl = 'stats.json';

const descDict = {
  concurrency: {
    name: 'concurrency',
    desc: 'the concurrency of pike',
  },
  sys: {
    name: 'sys memory',
    desc: 'the sys memory usage(MB)',
  },
  heapSys: {
    name: 'heap sys',
    desc: 'the heap sys memory usage(MB)',
  },
  heapInuse: {
    name: 'heap in use',
    desc: 'the heap in use memory usage(MB)',
  },
  routine: {
    name: 'go rountine',
    desc: 'the count of go rountine',
  },
  cacheCount: {
    name: 'cache count',
    desc: 'the cache count incliding cacheable and hit for pass',
  },
  cacheable: {
    name: 'cacheable count',
    desc: 'the cacheable count of cache',
  },
  hitForPass: {
    name: 'hit for pass count',
    desc: 'the hit for pass count of cache',
  },
  fetching: {
    name: 'fetching count',
    desc: 'the count of fetching request',
  },
  waiting: {
    name: 'waiting count',
    desc: 'the count of waiting request',
  },
};

const refreshStats = () => {
  fetch(statsUrl).then((res) => {
    return res.json();
  }).then((data) => {
    main.getStats(data);
  });
}

const actions = {
  getStats: (data) => state => {
    return data;
  },
}

const Pluzze = ({ name, count, desc, index }) => {
  if (count == null) {
    return;
  }
  const colors = [
    'green',
    'red',
    'purple',
    'yellow',
  ];
  const color = colors[index % colors.length];
  const cls = 'col-sm';
  return <li class={cls}>
    <div class={color}>
      <span>{name}</span>
      <h4
        onupdate={() => {
          chart.addData(name, count);
        }}
      >{count.toLocaleString()}</h4>
      <div class="chartWrapper">
        <canvas
          style='width:100%;height:130px'
          oncreate={element => {
            const ctx = element.getContext('2d');
            chart.init(ctx, name);
          }}
        ></canvas>
      </div>
      <p title={desc}>{desc}</p>
    </div>
  </li>
};

const StatsView = ({ state }) => {
  const keys = Object.keys(descDict);
  const list = keys.map((k, index) => {
    const item = descDict[k];
    const v = state[k];
    return <Pluzze name={item.name} count={v} index={index} desc={item.desc} />
  });
  return <ul class='row statsView'>
    {list}
  </ul>
}


let prevRequestCount = 0;
const RequestCountView = ({ state }) => {
  const k = 'requestCount';
  const name = 'request count';
  const desc = 'the count of request';
  if (prevRequestCount == 0) {
    prevRequestCount = state[k];
    return null;
  }
  const v = state[k] - prevRequestCount;
  prevRequestCount = state[k];
  return <ul class="statsView">
    <Pluzze name={name} count={v} index={0} desc={desc} />
  </ul>
}

const view = (state, actions) => (
  <div>
    <nav class="navBar">Pike Dashboard</nav>
    <div class="contentWrapper container">
      <div class="bkz">
        <h3 class="bla blb">QUICK PERFORMANCE</h3>
      </div>
      <StatsView state={state} />
      <div class="bkz">
        <h3 class="bla blb">REQUEST OF PER MINUTE</h3>
      </div>
      <RequestCountView state={state} />
    </div>
  </div>
)

const main = app(state, actions, view, document.body)

setInterval(refreshStats, 60 * 1000);
refreshStats();
