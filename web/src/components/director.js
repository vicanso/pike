import React from "react";
import request from "axios";
import { Spin, message, Icon } from "antd";

import { UPSTREAMS } from "../urls";
import "./director.sass";

function createList(data, key, name) {
  if (!data[key]) {
    return;
  }
  const value = data[key];
  let arr = null;
  if (Array.isArray(value)) {
    arr = value.map(item => {
      return <li key={item}>{item}</li>;
    });
  } else {
    const keys = Object.keys(value);
    arr = keys.map(item => {
      return <li key={item}>{`${item}:${value[item]}`}</li>;
    });
  }

  return (
    <div>
      <h5>{name || key}</h5>
      <ul>{arr}</ul>
    </div>
  );
}

class Director extends React.Component {
  state = {
    loading: false,
    upstreams: null
  };
  renderUpstreams() {
    const { loading, upstreams } = this.state;
    if (loading || !upstreams) {
      return;
    }
    if (upstreams.length === 0) {
      return (
        <div className="noUpstreams">
          <Icon type="info-circle" />
          There is no upstream.
        </div>
      );
    }
    return upstreams.map(item => {
      const servers = item.servers.map(server => {
        let icon = <Icon className="status" type="check-circle" />;
        if (server.status !== "healthy") {
          icon = <Icon className="status sick" type="close-circle" />;
        }
        return (
          <li key={server.url}>
            {server.url}
            {icon}
            {server.backup && <span className="backup">backup</span>}
          </li>
        );
      });

      return (
        <div className="upstream" key={item.name}>
          <h4>
            <div className="priority">priority:{item.priority}</div>
            {item.name}
            {item.policy && <span className="policy">{item.policy}</span>}
          </h4>
          <h5>servers</h5>
          <ul>{servers}</ul>
          {createList(item, "hosts")}
          {createList(item, "prefixs")}
          {createList(item, "rewrites")}
          {createList(item, "header")}
          {createList(item, "requestHeader", "request header")}
        </div>
      );
    });
  }
  render() {
    const { loading } = this.state;
    return (
      <div className="Director">
        {loading && (
          <div
            style={{
              textAlign: "center",
              paddingTop: "50px"
            }}
          >
            <Spin tip="Loading..." />
          </div>
        )}
        {this.renderUpstreams()}
      </div>
    );
  }
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await request.get(UPSTREAMS);
      this.setState({
        upstreams: data.upstreams
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        loading: false
      });
    }
  }
}

export default Director;
