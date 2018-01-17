import { h, app } from 'hyperapp';

import * as chart from './chart';


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



const Pluzze = ({ name, data, desc, index }) => {
  if (data == null || data.length === 0) {
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
  const count = data[data.length - 1].value;
  chart.setData(name, data);
  return <li class={cls}>
    <div class={color}>
      <span>{name}</span>
      <h4>{count.toLocaleString()}</h4>
      <div class="chartWrapper">
        <canvas
          style='width:100%;height:130px'
          oncreate={element => {
            const ctx = element.getContext('2d');
            chart.init(ctx, name);
            chart.setData(name, data);
          }}
          ondestroy={() => {
            chart.remove(name);
          }}
        ></canvas>
      </div>
      <p title={desc}>{desc}</p>
    </div>
  </li>
};

const StatsView = ({ state }) => {
  const performance = state.performance;
  if (!performance) {
    return;
  }
  const keys = Object.keys(descDict);
  const list = keys.map((k, index) => {
    const item = descDict[k];
    const v = performance[k];
    return <Pluzze name={item.name} data={v} index={index} desc={item.desc} />
  });
  return <ul class='row statsView'>
    {list}
  </ul>
}


const RequestCountView = ({ state }) => {
  const k = 'requestCount';
  const name = 'request count';
  const desc = 'the count of request';
  const performance = state.performance;
  if (!performance) {
    return;
  }
  const v = performance[k];
  return <ul class="statsView">
    <Pluzze name={name} data={v} index={0} desc={desc} />
  </ul>
}

const Performance = ({ state }) => {
  return <div
    class="performanceWrapper container"
  >
    <div class="bkz">
      <h3 class="bla blb">QUICK PERFORMANCE</h3>
    </div>
    <StatsView state={state} />
    <div class="bkz">
      <h3 class="bla blb">REQUEST COUNT</h3>
    </div>
    <RequestCountView state={state} />
  </div>
}

export default Performance
