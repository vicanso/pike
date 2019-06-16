import request from "axios";

import AddUpdateUpstream from "./add_update_upstream";
import { UPSTREAMS } from "../urls";
import { message } from "antd";

class UpdateUpstream extends AddUpdateUpstream {
  constructor() {
    super();
    this.state.type = "update";
  }
  submit(data) {
    const { name } = this.props.match.params;
    return request.patch(`${UPSTREAMS}/${name}`, data);
  }
  async componentDidMount() {
    const { name } = this.props.match.params;
    this.setState({
      spinning: true,
      spinTips: "Loading..."
    });
    try {
      const { data } = await request.get(`${UPSTREAMS}/${name}`);
      const updateData = {};
      if (data.name) {
        updateData.name = data.name;
      }

      if (data.policy) {
        const arr = data.policy.split(":");
        updateData.policy = arr[0];
        updateData.policyField = arr[1] || "";
      }
      if (data.ping) {
        updateData.ping = data.ping;
      }
      const fillHeaders = name => {
        if (!data[name]) {
          return;
        }
        const headers = [];
        Object.keys(data[name]).forEach(key => {
          const values = data[name][key];
          headers.push({
            key,
            value: values[0]
          });
        });
        updateData[name] = headers;
      };
      fillHeaders("requestHeader");
      fillHeaders("header");

      const fillArr = (name, updateName) => {
        if (!data[name] || data[name].length === 0) {
          return;
        }
        updateData[updateName || name] = data[name];
      };
      fillArr("servers", "backends");
      fillArr("prefixs");
      fillArr("hosts");
      fillArr("rewrites");

      this.setState(updateData);
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        inited: true,
        spinning: false
      });
    }
  }
}

export default UpdateUpstream;
