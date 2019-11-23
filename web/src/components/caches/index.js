import React from "react";
import axios from "axios";
import { message, Spin, Button } from "antd";
import _ from "lodash";

import ExForm from "../exform";
import ExTable from "../extable";
import i18n from "../../i18n";
import { CONFIGS, CONFIG } from "../../urls";
import "./caches.sass";

const category = "caches";

class Caches extends React.Component {
  state = {
    mode: "",
    loading: false,
    currentCache: null,
    caches: null
  };
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await axios.get(CONFIGS.replace(":category", category));
      this.setState({
        caches: data.caches
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
    const url = CONFIGS.replace(":category", category);
    try {
      await axios.post(url, data);
      const caches = this.state.caches.slice(0);
      const index = _.findIndex(caches, item => {
        return item.name === data.name;
      });
      if (index === -1) {
        caches.push(data);
      } else {
        caches[index] = data;
      }

      this.setState({
        mode: "",
        caches
      });
    } catch (err) {
      message.error(err.message);
      done();
    } finally {
    }
  }
  async handleDelete(item) {
    const url = CONFIG.replace(":category", category).replace(
      ":name",
      item.name
    );
    return axios.delete(url).then(() => {
      const caches = _.filter(this.state.caches, data => {
        return data.name !== item.name;
      });
      this.setState({
        caches
      });
    });
  }
  handleUpdate(item) {
    this.setState({
      mode: "update",
      currentCache: item
    });
  }
  renderConfigs() {
    const { loading, mode } = this.state;
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
    const columns = [
      {
        title: i18n("caches.name"),
        dataIndex: "name"
      },
      {
        title: i18n("caches.size"),
        dataIndex: "size",
        sorter: (a, b) => a.size - b.size
      },
      {
        title: i18n("caches.zone"),
        dataIndex: "zone",
        sorter: (a, b) => a.zone - b.zone
      },
      {
        title: i18n("caches.hitForPass"),
        dataIndex: "hitForPass",
        sorter: (a, b) => a.hitForPass - b.hitForPass
      },
      {
        title: i18n("common.description"),
        dataIndex: "description"
      }
    ];
    return (
      <ExTable
        onDelete={this.handleDelete.bind(this)}
        onUpdate={this.handleUpdate.bind(this)}
        rowKey={"name"}
        dataSource={this.state.caches}
        columns={columns}
      />
    );
  }
  render() {
    const fields = [
      {
        label: i18n("caches.name"),
        key: "name",
        placeholder: i18n("caches.namePlaceholder"),
        rules: [
          {
            required: true,
            message: i18n("caches.nameRequireMessage")
          }
        ]
      },
      {
        label: i18n("caches.zone"),
        key: "zone",
        type: "number",
        placeholder: i18n("caches.zonePlaceholder"),
        rules: [
          {
            required: true,
            message: i18n("caches.zoneRequireMessage")
          }
        ]
      },
      {
        label: i18n("caches.size"),
        key: "size",
        type: "number",
        placeholder: i18n("caches.sizePlaceholder"),
        rules: [
          {
            required: true,
            message: i18n("caches.sizeRequireMessage")
          }
        ]
      },
      {
        label: i18n("caches.hitForPass"),
        key: "hitForPass",
        type: "number",
        placeholder: i18n("caches.hitForPassPlaceholder"),
        rules: [
          {
            required: true,
            message: i18n("caches.hitForPassRequireMessage")
          }
        ]
      },
      {
        label: i18n("common.description"),
        key: "description",
        type: "textarea",
        placeholder: i18n("common.descriptionPlaceholder")
      }
    ];

    const { mode, loading, currentCache } = this.state;

    return (
      <div className="Caches">
        <Spin spinning={loading}>
          {this.renderConfigs()}
          {!mode && (
            <Button
              className="add"
              type="primary"
              onClick={() => {
                this.setState({
                  mode: "add",
                  currentCache: null
                });
              }}
            >
              {i18n("common.add")}
            </Button>
          )}
          <div className="form">
            {mode && (
              <ExForm
                originalData={currentCache}
                title={i18n("caches.createUpdateTitle")}
                fields={fields}
                description={i18n("caches.createUpdateDescription")}
                onSubmit={(value, done) => {
                  this.handleSubmit(value, done);
                }}
                onBack={() => {
                  this.setState({
                    currentCache: null,
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

export default Caches;
