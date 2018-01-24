import { h, app } from 'hyperapp';

let inputElement = null;
const AdminToken = ({ state }) => {
  return <div class="adminTokenWrapper container contentWrapper"><form>
    <div class="form-group">
      <label for="adminToken">Admin Token</label>
      <input
        type="text"
        class="form-control"
        id="adminToken"
        oncreate={(element) => {
          inputElement = element;
          element.focus();
        }}
        placeholder="Enter admin token"
      />
      <small class="form-text text-muted">The token you set in the pike's config</small>
    </div>
    <button
      type="submit"
      class="btn btn-primary"
      onclick={(element) => {
        localStorage.setItem('adminToken', inputElement.value)
        element.preventDefault();
        location.reload();
      }}
    >Submit</button>
  </form></div>
}

export default AdminToken
