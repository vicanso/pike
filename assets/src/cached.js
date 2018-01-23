import { h, app } from 'hyperapp';
import moment from 'moment';

import {
  getCacheds,
} from './actions';

const renderTable = (data) => {
  if (data.length === 0) {
    return <p class="tac">There is no cached data</p>;
  }
  const trList = data.map((item) => {
    return <tr>
      <td>{item.key}</td>
      <td>{item.ttl}</td>
      <td>{moment(item.createdAt * 1000).format('YYYY-MM-DD hh:mm:ss')}</td>
      <td>{moment((item.createdAt + item.ttl) * 1000).fromNow(true)}</td>
    </tr>
  });
  return <table class="table">
    <thead><tr>
      <th>Key</th>
      <th>TTL</th>
      <th>CreatedAt</th>
      <th>Expired</th>
    </tr></thead>
    <tbody>
      { trList }
    </tbody>
  </table>
}

const Cached = ({ state, actions, toggleCount }) => {
  return <div
    key={toggleCount}
    class="cachedWrapper container"
    oncreate={() => {
      getCacheds().then((data) => {
        actions.setCacheds(data);
      });
    }}
    ondestroy={() => {
      actions.resetCacheds();
    }}
  >
    <h3>Current Cached List</h5>
    {
      !state.cacheds && <p class="tac">Loading...</p>
    }
    {
      state.cacheds && renderTable(state.cacheds)
    }
  </div>
};

export default Cached;
