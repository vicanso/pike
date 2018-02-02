import { h, app } from 'hyperapp';
import moment from 'moment';

import {
  getCacheds,
  removeCache,
} from './actions';

import {
  createLineHeader,
} from './widget';

const renderTable = (data, keyword, actions) => {
  if (data.length === 0) {
    return <p class="tac">There is no cached data</p>;
  }
  const format = (ttl) => {
    if (ttl >= 3600) {
      return `${(ttl / 3600).toFixed(1)}h`;
    }
    if (ttl >= 60) {
      return `${(ttl / 60).toFixed(1)}m`;
    }
    return `${ttl}s`;
  }
  const trList = data.map((item) => {
    const {
      key,
      ttl,
      createdAt,
    } = item;
    if (keyword && key.indexOf(keyword) === -1) {
      return;
    }
    return <tr>
      <td class="key">{key}</td>
      <td class="ttl">{format(ttl)}</td>
      <td class="date">{moment(createdAt * 1000).format('YYYY-MM-DD hh:mm:ss')}</td>
      <td class="expired">{moment((createdAt + ttl) * 1000).fromNow(true)}</td>
      <td class="op">
        <a
          href="javascript:;"
          onclick={() => {
            removeCache(key).then(() => {
              getCacheds().then((data) => {
                actions.setCacheds(data);
              });
            });
          }}
        >DEL</a>
      </td>
    </tr>
  });
  return <table class="table">
    <thead><tr>
      <th class="key">Key</th>
      <th class="ttl">TTL</th>
      <th class="date">CreatedAt</th>
      <th class="expired">Expired</th>
      <th class="op">OP</th>
    </tr></thead>
    <tbody>
      { trList }
    </tbody>
  </table>
}

const Cached = ({ state, actions, toggleCount }) => {
  return <div
    key={toggleCount}
    class="cachedWrapper contentWrapper"
    oncreate={() => {
      getCacheds().then((data) => {
        actions.setCacheds(data);
      });
    }}
    ondestroy={() => {
      actions.resetCacheds();
    }}
  >
    { createLineHeader('Current Cached List') }
    <div class="keyFilter">
      <a
        href="javascript:;"
        onclick={() => actions.toggleFilter()}
      >filter</a>
      { state.showFilter && <input
        type="text"
        class="form-control"
        placeholder="Enter filter keyword"
        oncreate={e => e.focus()}
        oninput={e => actions.setFilterKeyword(e.target.value)}
      />
      }
    </div>
    {
      !state.cacheds && <p class="tac">Loading...</p>
    }
    {
      state.cacheds && renderTable(state.cacheds, state.filterKeyword, actions)
    }
  </div>
};

export default Cached;
