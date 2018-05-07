import _ from 'lodash';
import {mapState} from 'vuex';

import Chart from '../../components/chart';

const performanceItems = {
  concurrency: {
    name: 'concurrency',
    desc: 'the concurrency of pike',
  },
  requestCount: {
    name: 'requestCount',
    desc: 'the request count of pike',
  },
  cacheCount: {
    name: 'cache count',
    desc: 'the cache count incliding cacheable and hit for pass',
  },
  cacheable: {
    name: 'cacheable count',
    desc: 'the cacheable count of cache',
  },
  fetching: {
    name: 'fetching count',
    desc: 'the count of fetching request',
  },
  hitForPass: {
    name: 'hit for pass count',
    desc: 'the hit for pass count of cache',
  },
  waiting: {
    name: 'waiting count',
    desc: 'the count of waiting request',
  },
  heapSys: {
    name: 'heap sys',
    desc: 'the heap sys memory usage(MB)',
  },
  heapInuse: {
    name: 'heap in use',
    desc: 'the heap in use memory usage(MB)',
  },
  sys: {
    name: 'sys memory',
    desc: 'the sys memory usage(MB)',
  },
  routine: {
    name: 'go rountine',
    desc: 'the count of go rountine',
  },
  fileSize: {
    name: 'the size of db',
    desc: 'the data size of db',
  },
};

export default {
  data() {
    const colors = [
      'green',
      'red',
      'purple',
      'yellow',
    ];
    let index = 0;
    return {
      performanceItems: _.map(performanceItems, (item, key) => {
        const cls = {};
        const color = colors[index % colors.length];
        index += 1;
        cls[color] = true;
        item.cls = cls;
        item.key = key;
        return item
      }),
    };
  },
  components: {
    Chart,
  },
  methods: {
    getCount(name) {
      const value = _.last(this.performances);
      if (!value) {
        return 0;
      }
      return (value[name] || 0).toLocaleString();
    },
  },
  computed: {
    ...mapState({
      performances: ({pike}) => pike.performances,
    }),
  },
}