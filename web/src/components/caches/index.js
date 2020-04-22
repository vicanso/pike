import React from "react";
import axios from "axios";
import {
  message,
  Card,
  Form,
  Input,
  Row,
  Col,
  Button,
  Select,
  Spin,
  Table,
  Popconfirm,
  Icon
} from "antd";

import Configs from "../configs";
import { getCacheI18n, getCommonI18n } from "../../i18n";
import { CACHES } from "../../urls";
import "./caches.sass";

const category = "caches";
const { Option } = Select;
const oneSecond = 1000;
const oneMinute = 60 * oneSecond;
const oneHour = 60 * oneMinute;

const columns = [
  {
    title: getCacheI18n("name"),
    dataIndex: "name"
  },
  {
    title: getCacheI18n("size"),
    dataIndex: "size",
    sorter: (a, b) => a.size - b.size
  },
  {
    title: getCacheI18n("zone"),
    dataIndex: "zone",
    sorter: (a, b) => a.zone - b.zone
  },
  {
    title: getCacheI18n("hitForPass"),
    dataIndex: "hitForPass",
    sorter: (a, b) => a.hitForPass - b.hitForPass
  },
  {
    title: getCacheI18n("purgedAt"),
    dataIndex: "purgedAt",
    render: row => {
      if (!row) {
        return;
      }
      const arr = row.split(" ");
      return arr.map((item, index) => (
        <span
          style={{
            marginRight: "10px"
          }}
          key={index}
        >
          {item}
        </span>
      ));
    }
  },
  {
    title: getCommonI18n("description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: getCacheI18n("name"),
    key: "name",
    placeholder: getCacheI18n("namePlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("size"),
    key: "size",
    type: "number",
    placeholder: getCacheI18n("sizePlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("sizeRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("zone"),
    key: "zone",
    type: "number",
    placeholder: getCacheI18n("zonePlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("zoneRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("hitForPass"),
    key: "hitForPass",
    type: "number",
    placeholder: getCacheI18n("hitForPassPlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("hitForPassRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("purgedAt"),
    key: "purgedAt",
    placeholder: getCacheI18n("purgedAtPlaceholder")
  },
  {
    label: getCommonI18n("description"),
    key: "description",
    type: "textarea",
    placeholder: getCommonI18n("descriptionPlaceholder")
  }
];

class Caches extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      currentCache: "",
      limit: 50,
      keyword: "",
      caches: null,
      processing: false,
      title: getCacheI18n("createUpdateTitle"),
      description: getCacheI18n("createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
  async listCaches() {
    const { currentCache, limit, processing, configs, keyword } = this.state;
    if (processing) {
      return;
    }
    if (!limit) {
      message.error("limit must be required");
      return;
    }
    this.setState({
      processing: true
    });
    try {
      const name = currentCache || configs[0].name;
      const { data } = await axios.get(CACHES.replace(":name", name), {
        params: {
          limit,
          keyword
        }
      });
      this.setState({
        caches: data
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        processing: false
      });
    }
  }
  async handleDelete(item) {
    const { processing, currentCache, configs, caches } = this.state;
    if (processing) {
      return;
    }
    this.setState({
      processing: true
    });
    try {
      const name = currentCache || configs[0].name;
      await axios.delete(CACHES.replace(":name", name), {
        params: {
          key: item.key
        }
      });
      const arr = [];
      caches.forEach(cache => {
        if (cache.key !== item.key) {
          arr.push(cache);
        }
      });
      this.setState({
        caches: arr
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        processing: false
      });
    }
  }
  renderCachesTable() {
    const { processing, caches } = this.state;
    const columns = [
      {
        title: getCacheI18n("key"),
        dataIndex: "key"
      },
      {
        title: getCacheI18n("createdAt"),
        dataIndex: "createdAt",
        width: 220,
        sorter: (a, b) => a.createdAt - b.createdAt,
        render: row => {
          return new Date(row * 1000).toLocaleString();
        }
      },
      {
        title: getCacheI18n("expiredAt"),
        dataIndex: "expiredAt",
        width: 220,
        sorter: (a, b) => a.expiredAt - b.expiredAt,
        render: row => {
          const value = row * 1000;
          const now = Date.now();
          const ms = value - now;
          if (ms < oneMinute) {
            return `${Math.round(ms / oneSecond)}s`;
          }
          if (ms < oneHour) {
            return `${Math.round(ms / oneMinute)}m`;
          }
          return new Date(value).toLocaleString();
        }
      },
      {
        title: getCommonI18n("action"),
        width: 100,
        render: row => {
          return (
            <Popconfirm
              key="ondelete"
              title={getCommonI18n("deleteTips")}
              onConfirm={() => {
                this.handleDelete(row);
              }}
            >
              <a
                href="/delete"
                onClick={e => {
                  e.preventDefault();
                }}
              >
                <Icon type="delete" />
                {getCommonI18n("delete")}
              </a>
            </Popconfirm>
          );
        }
      }
    ];
    return (
      <Spin spinning={processing}>
        <Table rowKey={"key"} dataSource={caches} columns={columns} />
      </Spin>
    );
  }
  renderCaches() {
    const { configs, limit } = this.state;
    if (!configs || configs.length === 0) {
      return;
    }
    const formItemLayout = {
      labelCol: {
        xs: { span: 24 },
        sm: { span: 5 }
      },
      wrapperCol: {
        xs: { span: 24 },
        sm: { span: 19 }
      }
    };
    const tailFormItemLayout = {
      wrapperCol: {
        xs: {
          span: 24,
          offset: 0
        },
        sm: {
          span: 18,
          offset: 4
        }
      }
    };
    const cacheSelectItems = configs.map(item => {
      return (
        <Option key={item.name} value={item.name}>
          {item.name}
        </Option>
      );
    });
    return (
      <Card className="cacheList" title={getCacheI18n("caches")}>
        <Form
          {...formItemLayout}
          onSubmit={e => {
            e.preventDefault();
            this.listCaches();
          }}
        >
          <Row gutter={8}>
            <Col span={7}>
              <Form.Item label={getCacheI18n("name")}>
                <Select
                  defaultValue={configs[0].name}
                  onChange={value => {
                    this.setState({
                      currentCache: value
                    });
                  }}
                >
                  {cacheSelectItems}
                </Select>
              </Form.Item>
            </Col>
            <Col span={7}>
              <Form.Item label={getCacheI18n("limit")}>
                <Input
                  type="number"
                  defaultValue={limit}
                  onChange={e => {
                    this.setState({
                      limit: e.target.valueAsNumber
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={7}>
              <Form.Item label={getCacheI18n("keyword")}>
                <Input
                  type="text"
                  allowClear
                  onChange={e => {
                    this.setState({
                      keyword: e.target.value.trim()
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={3}>
              <Form.Item key="submit" {...tailFormItemLayout}>
                <Button type="primary" htmlType="submit" className="search">
                  {getCommonI18n("search")}
                </Button>
              </Form.Item>
            </Col>
          </Row>
        </Form>
        {this.renderCachesTable()}
      </Card>
    );
  }
  render() {
    return (
      <div className="Caches">
        <div>{super.render()}</div>
        {this.renderCaches()}
      </div>
    );
  }
}

export default Caches;
