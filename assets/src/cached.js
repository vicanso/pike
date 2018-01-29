import { h, app } from 'hyperapp';
import moment from 'moment';

import {
  getCacheds,
  removeCache,
} from './actions';

import {
  createLineHeader,
} from './widget';

const renderTable = (data, actions) => {
  if (data.length === 0) {
    return <p class="tac">There is no cached data</p>;
  }
  const trList = data.map((item) => {
    return <tr>
      <td class="key">{item.key}</td>
      <td>{item.ttl}</td>
      <td>{moment(item.createdAt * 1000).format('YYYY-MM-DD hh:mm:ss')}</td>
      <td>{moment((item.createdAt + item.ttl) * 1000).fromNow(true)}</td>
      <td>
        <a
          href="javascript:;"
          onclick={() => {
            removeCache(item.key).then(() => {
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
      <th>TTL</th>
      <th>CreatedAt</th>
      <th>Expired</th>
      <th>OP</th>
    </tr></thead>
    <tbody>
      { trList }
    </tbody>
  </table>
}

const Cached = ({ state, actions, toggleCount }) => {
  return <div
    key={toggleCount}
    class="cachedWrapper container contentWrapper"
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
    {
      !state.cacheds && <p class="tac">Loading...</p>
    }
    {
      state.cacheds && renderTable(state.cacheds, actions)
    }
  </div>
};

export default Cached;
