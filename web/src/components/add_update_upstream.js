import React from "react";
import {
  Form,
  Input,
  Select,
  Button,
  Row,
  Col,
  Icon,
  message,
  Switch,
  Spin
} from "antd";

import "./add_update_upstream.sass";

const Option = Select.Option;

const needFiledPolicyList = ["header", "cookie"];

class AddUpdateUpstream extends React.Component {
  state = {
    inited: false,
    type: "",
    spinning: false,
    spinTips: "",
    name: "",
    policy: "",
    policyField: "",
    ping: "",
    prefixs: [],
    hosts: [],
    rewrites: [],
    backends: [],
    responseHeader: [],
    requestHeader: []
  };
  async handleSubmit(e) {
    const { history } = this.props;
    e.preventDefault();
    // 点击时，select的数据有可能未更新，延时操作
    await new Promise(resolve => {
      setTimeout(resolve, 100);
    });
    const {
      name,
      policy,
      ping,
      policyField,
      prefixs,
      hosts,
      rewrites,
      backends,
      responseHeader,
      requestHeader
    } = this.state;

    this.setState({
      spinning: true,
      spinTips: "submitting"
    });
    try {
      if (!name || !backends.length) {
        throw new Error("name and backends can't be null");
      }
      if (name.indexOf(".") !== -1) {
        throw new Error("name can't include '.'");
      }
      const backendList = [];
      backends.forEach(item => {
        if (!item || !item.url) {
          return;
        }
        const reg = new RegExp("^http(s)?://[a-zA-Z0-9][-a-zA-Z0-9]{0,62}");
        if (!reg.test(item.url)) {
          throw new Error(`backend(${item.url}) is invalid`);
        }
        let url = item.url;
        if (item.backup) {
          url += "|backup";
        }
        backendList.push(url);
      });

      if (needFiledPolicyList.includes(policy) && !policyField) {
        throw new Error("the field of policy can't be null");
      }

      const data = {
        name,
        backends: backendList
      };
      if (policy) {
        data.policy = policy;
      }
      if (needFiledPolicyList.includes(policy)) {
        data.policy = `${data.policy}:${policyField}`;
      }

      if (ping) {
        data.ping = ping;
      }

      if (prefixs.length) {
        data.prefixs = prefixs;
      }
      if (hosts.length) {
        data.hosts = hosts;
      }
      if (rewrites.length) {
        data.rewrites = rewrites;
      }
      const filterResponseHeader = responseHeader.filter(
        item => item && item.key && item.value
      );
      if (filterResponseHeader.length) {
        data.responseHeader = filterResponseHeader.map(
          item => `${item.key}:${item.value}`
        );
      }
      const filterRequestHeader = requestHeader.filter(
        item => item && item.key && item.value
      );
      if (filterRequestHeader.length) {
        data.requestHeader = filterRequestHeader.map(
          item => `${item.key}:${item.value}`
        );
      }
      await this.submit(data);
      message.info("add/update upstream successful");
      await new Promise(resolve => {
        setTimeout(resolve, 500);
      });
      if (history) {
        history.goBack();
      }
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        spinning: false
      });
    }
  }
  renderPolicySelector() {
    const { policy, policyField } = this.state;
    const policyList = [
      "first",
      "random",
      "roundRobin",
      "leastconn",
      "ipHash",
      "header",
      "cookie"
    ].map(item => {
      return <Option key={item}>{item}</Option>;
    });
    let input = null;
    if (needFiledPolicyList.includes(policy)) {
      input = (
        <Input
          type="text"
          defaultValue={policyField}
          onChange={e => {
            this.setState({
              policyField: e.target.value
            });
          }}
          placeholder={`Input the field of ${policy}`}
        />
      );
    }
    return (
      <Form.Item label="Policy">
        <Select
          defaultValue={policy}
          placeholder="Select the policy for upstream"
          onChange={value => {
            this.setState({
              policy: value
            });
          }}
        >
          {policyList}
        </Select>
        {input}
      </Form.Item>
    );
  }
  renderHeader(name) {
    const values = this.state[name] || [];
    const getRow = (headerValue, index) => {
      if (!headerValue) {
        return null;
      }
      const { key, value } = headerValue;
      return (
        <Row key={`${name}-${index}`}>
          <Col span={11}>
            <Input
              type="text"
              placeholder="Input header's name"
              defaultValue={key}
              onChange={e => {
                const arr = values.slice(0);
                if (!arr[index]) {
                  arr[index] = {};
                }
                arr[index].key = e.target.value.trim();
                const data = {};
                data[name] = arr;
                this.setState(data);
              }}
            />
          </Col>
          <Col className="divide" span={1}>
            <span>:</span>
          </Col>
          <Col span={11}>
            <Input
              type="text"
              placeholder="Input header's value"
              defaultValue={value}
              onChange={e => {
                const arr = values.slice(0);
                if (!arr[index]) {
                  arr[index] = {};
                }
                arr[index].value = e.target.value.trim();
                const data = {};
                data[name] = arr;
                this.setState(data);
              }}
            />
          </Col>
          <Col className="remove" span={1}>
            <a
              href="/remove-header"
              onClick={e => {
                e.preventDefault();
                const arr = values.slice(0);
                // 为了保证数组的顺序，直接置空，不删除
                arr[index] = null;
                const data = {};
                data[name] = arr;
                this.setState(data);
              }}
            >
              <Icon type="close-circle" />
            </a>
          </Col>
        </Row>
      );
    };

    const rows = [];
    values.forEach((item, index) => {
      rows.push(getRow(item, index));
    });
    rows.push(getRow({}, rows.length));
    return <div>{rows}</div>;
  }
  renderPrefixs() {
    const { prefixs } = this.state;
    const arr = prefixs.map(item => <Option key={item}>{item}</Option>);
    return (
      <Form.Item label="Prefixs">
        <Select
          defaultValue={prefixs}
          placeholder="Input the prefix of upstream, e.g.: /api"
          mode="tags"
          onChange={value => {
            this.setState({
              prefixs: value
            });
          }}
        >
          {arr}
        </Select>
      </Form.Item>
    );
  }
  renderHosts() {
    const { hosts } = this.state;
    const arr = hosts.map(item => <Option key={item}>{item}</Option>);
    return (
      <Form.Item label="Hosts">
        <Select
          defaultValue={hosts}
          placeholder="Input the hosts of upstream, e.g.: aslant.site"
          mode="tags"
          onChange={value => {
            this.setState({
              hosts: value
            });
          }}
        >
          {arr}
        </Select>
      </Form.Item>
    );
  }
  renderWrites() {
    const { rewrites } = this.state;
    const arr = rewrites.map(item => <Option key={item}>{item}</Option>);
    return (
      <Form.Item label="Rewrites">
        <Select
          defaultValue={rewrites}
          placeholder="Input the rewrite of upstream, e.g.: /api/*:/$1"
          mode="tags"
          onChange={value => {
            this.setState({
              rewrites: value
            });
          }}
        >
          {arr}
        </Select>
      </Form.Item>
    );
  }
  renderBackends() {
    const { backends } = this.state;
    const arr = [];

    const appendBackend = (item, index) => {
      if (!item) {
        return;
      }
      arr.push(
        <div key={`backend-${index}`}>
          <Row gutter={8}>
            <Col span={20}>
              <Input
                defaultValue={item.url}
                type="text"
                placeholder="Input the url of backend, e.g.: http://127.0.0.1:3000"
                onChange={e => {
                  const arr = this.state.backends.slice(0);
                  if (!arr[index]) {
                    arr[index] = {};
                  }
                  arr[index].url = e.target.value;
                  this.setState({
                    backends: arr
                  });
                }}
              />
            </Col>
            <Col span={3}>
              <Switch
                defaultChecked={item.backup}
                onChange={value => {
                  const arr = this.state.backends.slice(0);
                  if (!arr[index]) {
                    arr[index] = {};
                  }
                  arr[index].backup = value;
                  this.setState({
                    backends: arr
                  });
                }}
                checkedChildren="backup"
                unCheckedChildren="backup"
              />
            </Col>
            <Col className="remove" span={1}>
              <a
                href="/remove-backend"
                onClick={e => {
                  e.preventDefault();
                  const arr = this.state.backends.slice(0);
                  // 为了保证数组的顺序，直接置空，不删除
                  arr[index] = null;
                  this.setState({
                    backends: arr
                  });
                }}
              >
                <Icon type="close-circle" />
              </a>
            </Col>
          </Row>
        </div>
      );
    };

    backends.forEach(appendBackend);
    appendBackend({}, backends.length);
    return <Form.Item label="Backends">{arr}</Form.Item>;
  }
  renderForm() {
    const { inited, name, ping, type } = this.state;

    if (!inited) {
      return <div />;
    }
    const formItemLayout = {
      labelCol: {
        xs: { span: 24 },
        sm: { span: 4 }
      },
      wrapperCol: {
        xs: { span: 24 },
        sm: { span: 20 }
      }
    };
    const tailFormItemLayout = {
      wrapperCol: {
        xs: {
          span: 24,
          offset: 0
        },
        sm: {
          span: 22,
          offset: 2
        }
      }
    };
    return (
      <Form {...formItemLayout} onSubmit={this.handleSubmit.bind(this)}>
        <Form.Item label="Name">
          <Input
            defaultValue={name}
            type="text"
            onChange={e => {
              this.setState({
                name: e.target.value
              });
            }}
            placeholder="Input the name of upstream"
          />
        </Form.Item>
        {this.renderBackends()}
        {this.renderPolicySelector()}
        <Form.Item label="Ping">
          <Input
            defaultValue={ping}
            type="text"
            placeholder="Input the health check url, e.g.: /ping"
            onChange={e => {
              this.setState({
                ping: e.target.value
              });
            }}
          />
        </Form.Item>
        <Form.Item label="Request Header">
          {this.renderHeader("requestHeader")}
        </Form.Item>
        <Form.Item label="Response Header">
          {this.renderHeader("responseHeader")}
        </Form.Item>
        {this.renderPrefixs()}
        {this.renderHosts()}
        {this.renderWrites()}
        <Form.Item {...tailFormItemLayout}>
          <Button className="submit" type="primary" htmlType="submit">
            {type.toUpperCase()}
          </Button>
        </Form.Item>
      </Form>
    );
  }
  render() {
    const { spinning, spinTips } = this.state;
    return (
      <div className="AddUpdateUpstream">
        <Spin spinning={spinning} tip={spinTips}>
          {this.renderForm()}
        </Spin>
      </div>
    );
  }
}

export default AddUpdateUpstream;
