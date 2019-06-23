import React from "react";
import request from "axios";
import { Link } from "react-router-dom";
import { Spin, message, Icon, Card, Col, Row } from "antd";

import { CONFIGS } from "../urls";
import { UPDATE_CONFIG_PATH } from "../paths";
import "./config.sass";

class Config extends React.Component {
  state = {
    loading: false,
    basicConfig: null,
    applicationInfo: null,
    directorConfig: null
  };
  renderConfig() {
    const { basicConfig, directorConfig, applicationInfo } = this.state;
    const updateBasicPath = UPDATE_CONFIG_PATH.replace(":name", "basic");

    const cols = [
      {
        name: "Go Version",
        key: "goVersion"
      },
      {
        name: "Started At",
        key: "startedAt"
      },
      {
        name: "Builded At",
        key: "buildedAt"
      },
      {
        name: "Version",
        key: "version"
      },
      {
        name: "CommitId",
        key: "commitId"
      }
    ];
    const arr = cols.map(item => {
      if (!applicationInfo) {
        return null;
      }
      return (
        <Col span={8} className="info" key={item.key}>
          <span className="name">{item.name}</span>
          <span className="value">{applicationInfo[item.key]}</span>
        </Col>
      );
    });

    return (
      <div>
        <Card size="small" title="Application Information">
          <Row>{arr}</Row>
        </Card>
        <Card
          size="small"
          title={
            <div>
              Basic Config
              <Link
                className="update"
                to={updateBasicPath}
                title="update the config"
              >
                <Icon type="edit" />
              </Link>
            </div>
          }
        >
          <pre>{basicConfig}</pre>
        </Card>
        <Card size="small" title="Director Config">
          <pre>{directorConfig}</pre>
        </Card>
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
        {!loading && this.renderConfig()}
      </div>
    );
  }
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await request.get(CONFIGS);
      const keys = ["buildedAt", "startedAt"];
      keys.forEach(key => {
        const v = data.applicationInfo[key];
        if (v) {
          data.applicationInfo[key] = new Date(v).toLocaleString();
        }
      });
      this.setState({
        applicationInfo: data.applicationInfo,
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
