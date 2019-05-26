import React from "react";
import request from "axios";

import { Spin, message } from "antd";

import { CONFIGS } from "../urls";
import "./config.sass";

class Config extends React.Component {
  state = {
    loading: false,
    basicConfig: null,
    directorConfig: null
  };
  renderConfig() {
    const { basicConfig, directorConfig } = this.state;

    return (
      <div>
        <div className="yaml">
          <h3>Basic Config</h3>
          <pre>{basicConfig}</pre>
        </div>
        <div className="yaml">
          <h3>Director Config</h3>
          <pre>{directorConfig}</pre>
        </div>
      </div>
    );
  }
  render() {
    const { loading } = this.state;
    return (
      <div className="Config">
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
        {this.renderConfig()}
      </div>
    );
  }
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await request.get(CONFIGS);
      this.setState({
        basicConfig: data.basicYaml,
        directorConfig: data.directorYaml
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

export default Config;
