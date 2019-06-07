import request from "axios";

import AddUpdateUpstream from "./add_update_upstream";
import { UPSTREAMS } from "../urls";

class AddUpstream extends AddUpdateUpstream {
  constructor() {
    super();
    this.state.type = "add";
  }
  submit(data) {
    request.post(UPSTREAMS, data);
  }
}

export default AddUpstream;
