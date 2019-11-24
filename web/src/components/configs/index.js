import React from "react";
import axios from "axios";
import { message, Spin, Button } from "antd";
import _ from "lodash";

import ExForm from "../exform";
import ExTable from "../extable";
import i18n from "../../i18n";
import { CONFIGS, CONFIG } from "../../urls";
import "./configs.sass";

class Configs extends React.Component {
  state = {
    category: "",
    mode: "",
    title: "",
    description: "",
    loading: false,
    currentConfig: null,
    columns: null,
    fields: null,
    configs: null
  };
  async componentDidMount() {
    const { category, loading } = this.state;
    if (loading) {
      return;
    }
    this.setState({
      loading: true
    });
    try {
      const { data } = await axios.get(CONFIGS.replace(":category", category));
      this.setState({
        configs: data[category]
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        loading: false
      });
    }
  }
  async handleSubmit(data, done) {
    const { category } = this.state;
    const url = CONFIGS.replace(":category", category);
    try {
      await axios.post(url, data);
      const configs = this.state.configs.slice(0);
      const index = _.findIndex(configs, item => {
        return item.name === data.name;
      });
      if (index === -1) {
        configs.push(data);
      } else {
        configs[index] = data;
      }

      this.setState({
        mode: "",
        configs
      });
    } catch (err) {
      message.error(err.message);
      done();
    } finally {
    }
  }
  async handleDelete(item) {
    const { category } = this.state;
    const url = CONFIG.replace(":category", category).replace(
      ":name",
      item.name
    );
    return axios.delete(url).then(() => {
      const configs = _.filter(this.state.configs, data => {
        return data.name !== item.name;
      });
      this.setState({
        configs
      });
    });
  }
  handleUpdate(item) {
    this.setState({
      mode: "update",
      currentConfig: item
    });
  }
  renderConfigs() {
    const { loading, mode, columns, configs } = this.state;
    // 如果其它模式下，则不输出列表
    if (mode) {
      return;
    }
    if (loading) {
      return (
        <div
          style={{
            height: "300px"
          }}
        ></div>
      );
    }

    return (
      <ExTable
        onDelete={this.handleDelete.bind(this)}
        onUpdate={this.handleUpdate.bind(this)}
        rowKey={"name"}
        dataSource={configs}
        columns={columns}
      />
    );
  }
  render() {
    const {
      mode,
      loading,
      currentConfig,
      fields,
      title,
      description
    } = this.state;

    return (
      <div className="Configs">
        <Spin spinning={loading}>
          {this.renderConfigs()}
          {!mode && (
            <Button
              className="add"
              type="primary"
              onClick={() => {
                this.setState({
                  mode: "add",
                  currentConfig: null
                });
              }}
            >
              {i18n("common.add")}
            </Button>
          )}
          <div className="form">
            {mode && (
              <ExForm
                originalData={currentConfig}
                title={title}
                fields={fields}
                description={description}
                onSubmit={(value, done) => {
                  this.handleSubmit(value, done);
                }}
                onBack={() => {
                  this.setState({
                    currentConfig: null,
                    mode: ""
                  });
                }}
              />
            )}
          </div>
        </Spin>
      </div>
    );
  }
}

export default Configs;
