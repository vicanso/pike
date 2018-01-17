import { h, app } from 'hyperapp';

import {
  getBlackIPs,
  addBlackIP,
} from './actions';

let inputElement = null;
const BlackIP = ({ state, actions, toggleCount }) => {
  const refeshBlackIPList = () => {
    getBlackIPs().then((data) => {
      actions.setBlackIPList(data);
    });
  };
  return <div
    class="blackIPWrapper container"
    key={toggleCount}
    oncreate={refeshBlackIPList}
  >
    <form>
      <div class="form-group">
        <label for="blackIP">Black IP</label>
        <input
          type="text"
          class="form-control"
          id="blackIP"
          oncreate={(element) => {
            inputElement = element;
          }}
          placeholder="Enter the black ip"
        />
        <small class="form-text text-muted">The ip will be blocked</small>
      </div>
      <button
        type="submit"
        class="btn btn-primary"
        onclick={(element) => {
          const value = inputElement.value;
          console.dir(value);
          if (value) {
            addBlackIP(value).then(refeshBlackIPList);
          }
          element.preventDefault();
        }}
      >Submit</button>
    </form>
  </div>
};

export default BlackIP;
