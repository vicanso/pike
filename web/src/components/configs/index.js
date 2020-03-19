import React from "react";
import axios from "axios";
import { message, Spin, Button } from "antd";

import ExForm from "../exform";
import ExTable from "../extable";
import { getCommonI18n } from "../../i18n";
import { CONFIGS, CONFIG } from "../../urls";
import "./configs.sass";

class Configs extends React.Component {
  state = {
    single: false,
    disabledDelete: false,
    category: "",
    mode: "",
    title: "",
    description: "",
    minWidth: 0,
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
      let configs = data[category];
      if (!Array.isArray(configs)) {
        if (Object.keys(configs).length === 0) {
          configs = [];
        } else {
          configs = [configs];
        }
      }
      configs.forEach((item, index) => {
        if (!item.name) {
          item.name = `${index}`;
        }
      });
      this.setState({
        configs
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
    const { category, single } = this.state;
    const url = CONFIGS.replace(":category", category);
    try {
      await axios.post(url, data);
      const configs = this.state.configs.slice(0);
      if (!single) {
        let index = -1;
        configs.forEach((item, i) => {
          if (item.name === data.name) {
            index = i;
          }
        });
        if (index === -1) {
          configs.push(data);
        } else {
          configs[index] = data;
        }
      } else {
        configs[0] = data;
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
      const configs = this.state.configs.filter(data => {
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
    const {
      loading,
      mode,
      columns,
      configs,
      disabledDelete,
      minWidth
    } = this.state;
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
    const onDelete = disabledDelete ? null : this.handleDelete.bind(this);

    return (
      <ExTable
        minWidth={minWidth}
        onDelete={onDelete}
        onUpdate={this.handleUpdate.bind(this)}
        rowKey={"name"}
        dataSource={configs}
        columns={columns}
      />
    );
  }
  renderAdd() {
    const { mode, single, configs } = this.state;
    if (mode || (single && configs && configs.length !== 0)) {
      return;
    }
    return (
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
        {getCommonI18n("add")}
      </Button>
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
          {this.renderAdd()}
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
