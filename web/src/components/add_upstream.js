import request from "axios";

import AddUpdateUpstream from "./add_update_upstream";
import { UPSTREAMS } from "../urls";

class AddUpstream extends AddUpdateUpstream {
  constructor() {
    super();
    this.state.type = "add";
    this.state.inited = true;
  }
  submit(data) {
    return request.post(UPSTREAMS, data);
  }
}

export default AddUpstream;
