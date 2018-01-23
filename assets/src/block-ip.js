import { h, app } from 'hyperapp';

import {
  getBlockIPs,
  addBlockIP,
  removeBlockIP,
} from './actions';

let inputElement = null;

const createBlockIPTable = (actions, data) => {
  if (!data || !data.ipList || data.ipList.length === 0) {
    return null;
  }
  const ipList = data.ipList;
  const trList = ipList.map((item) => {
    return <tr>
      <td>{item}</td>
      <td title="remove the ip address">
        <a
          class="redColor"
          href="javascript:;"
          onclick={() => {
            removeBlockIP(item).then(() => {
              getBlockIPs().then((data) => {
                actions.setBlockIPList(data);
              });
            });
          }}
        >rm</a>
      </td>
    </tr>
  });
  return <div class="">
    <h3>Current Block IP List</h5>
    <table class="table">
      <thead><tr>
        <th>IP</th>
        <th>OP</th>
      </tr></thead>
      <tbody>
        { trList }
      </tbody>
    </table>
  </div>
}

const BlockIP = ({ state, actions, toggleCount }) => {
  const refeshBlockIPList = () => {
    getBlockIPs().then((data) => {
      actions.setBlockIPList(data);
    });
  };
  const blockIPList = state.blockIPList;
  return <div
    class="blockIPWrapper container"
    key={toggleCount}
    oncreate={refeshBlockIPList}
  >
    { createBlockIPTable(actions, blockIPList) }
    <h5>Create New Block IP</h5>
    <form>
      <div class="form-group">
        <label for="blockIP">Block IP</label>
        <input
          type="text"
          class="form-control"
          id="blockIP"
          oncreate={(element) => {
            inputElement = element;
          }}
          placeholder="Enter the block ip"
        />
        <small class="form-text text-muted">The ip will be blocked</small>
      </div>
      <button
        type="submit"
        class="btn btn-primary btn-block"
        onclick={(element) => {
          const value = inputElement.value;
          if (value) {
            addBlockIP(value).then(refeshBlockIPList);
          }
          element.preventDefault();
        }}
      >Submit</button>
    </form>
  </div>
};

export default BlockIP;
