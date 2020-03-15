import React from "react";
import axios from "axios";
import { message, Card, Row, Col, Spin } from "antd";

import i18n from "../../i18n";
import { APPLICATION } from "../../urls";
import "./home.sass";
import { toLocalTime } from "../../util";

const getAppI18n = name => i18n("application." + name);

class Home extends React.Component {
  state = {
    informations: null,
    loading: true
  };
  async componentDidMount() {
    try {
      const { data } = await axios.get(APPLICATION);
      data.buildedAt = toLocalTime(data.buildedAt);
      data.startedAt = toLocalTime(data.startedAt);
      this.setState({
        informations: data
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        loading: false
      });
    }
  }
  renderInformations() {
    const { informations } = this.state;
    if (!informations) {
      return;
    }
    const keys = [
      "buildedAt",
      "startedAt",
      "version",
      "goos",
      "maxProcs",
      "numGoroutine"
    ];
    const arr = keys.map(key => {
      return (
        <Col className="basicInfos" span={8} key={key}>
          <span>{getAppI18n(key)}</span>
          {informations[key]}
        </Col>
      );
    });
    return <Row gutter={16}>{arr}</Row>;
  }
  render() {
    const { loading } = this.state;
    return (
      <div className="Home">
        <Card title={getAppI18n("title")}>
          <Spin spinning={loading}>{this.renderInformations()}</Spin>
        </Card>
      </div>
    );
  }
}

export default Home;
