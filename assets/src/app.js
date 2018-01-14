import { h, app } from 'hyperapp';
import 'whatwg-fetch';
import 'chart.js/dist/Chart.js'

import './global.sss'
import './app.sss'


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

let chart = null;

const refreshStats = () => {
  fetch('/stats.json').then((res) => {
    return res.json();
  }).then((data) => {
    main.getStats(data);
    if (chart) {
      chart.config.data.datasets[0].data = data.requestCountList;
      chart.update();
    }
  });
}

const actions = {
  getStats: (data) => state => {
    return data;
  },
}

const getRuquesetCountChart = (ctx) => {
  const labels = [];
  for (let i = 0; i < 24; i++)  {
    for(let j = 0; j < 12; j++) {
      let hour = i;
      if (hour < 10) {
        hour = `0${hour}`;
      }
      let minute = j * 5;
      if (minute < 10) {
        minute = `0${minute}`;
      }
      labels.push(`${hour}:${minute}`);
    }
  }
  const config = {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label: "Request count per minute",
        backgroundColor: '#e74759',
        borderColor: '#e74759',
        data: [
        ],
        fill: false,
      }]
    },
    options: {
      responsive: true,
      title: {
        display: true,
        text: 'Pike Request Count'
      },
      tooltips: {
        mode: 'index',
        intersect: false,
      },
      hover: {
        mode: 'nearest',
        intersect: true
      },
      scales: {
        xAxes: [{
          display: true,
          scaleLabel: {
            display: true,
            labelString: 'Time'
          }
        }],
        yAxes: [{
          display: true,
          scaleLabel: {
            display: true,
            labelString: 'Count'
          }
        }]
      }
    }
  };
  return new Chart(ctx, config);
}

const Pluzze = ({ name, count, desc, index }) => {
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
      <h4>{count.toLocaleString()}</h4>
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
      <canvas
        oncreate={element => {
          const ctx = element.getContext('2d');
          chart = getRuquesetCountChart(ctx);
        }}
      >
      </canvas>
    </div>
  </div>
)

const main = app(state, actions, view, document.body)

setInterval(refreshStats, 5000);
// setInterval(() => {
// }, 1000);
// main.getStats();