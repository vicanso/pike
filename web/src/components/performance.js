import React from "react";
import request from "axios";
import { Spin, message, Row, Col, Card } from "antd";

import "./performance.sass";
import { STATS } from "../urls";

class Performance extends React.Component {
  state = {
    loading: false,
    stats: null
  };
  renderStats() {
    const { stats } = this.state;
    if (!stats) {
      return;
    }
    const cols = [
      {
        name: "Concurrency",
        key: "concurrency"
      },
      {
        name: "MaxProcs",
        key: "goMaxProcs"
      },
      {
        name: "Go Routine",
        key: "routine"
      },
      {
        name: "Memory(sys)",
        key: "sys"
      },
      {
        name: "Memory(heapSys)",
        key: "heapSys"
      },
      {
        name: "Memory(heapInuse)",
        key: "heapInuse"
      },
      {
        name: "Version",
        key: "version"
      },
      {
        name: "Go Version",
        key: "goVersion"
      }
    ];
    const arr = cols.map(item => {
      return (
        <Col span={8} className="info" key={item.key}>
          <span className="name">{item.name}</span>
          <span className="value">{stats[item.key]}</span>
        </Col>
      );
    });
    return (
      <Card title="Basic Informations">
        <Row>{arr}</Row>
      </Card>
    );
  }
  render() {
    const { loading } = this.state;
    return (
      <div className="Performance">
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
        {this.renderStats()}
      </div>
    );
  }
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await request.get(STATS);
      this.setState({
        stats: data.stats
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

export default Performance;
