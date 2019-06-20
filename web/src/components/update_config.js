import React from "react";
import request from "axios";
import { Spin, Form, Card, message, Input, Select, Button, Switch } from "antd";

import "./update_config.sass";
import { CONFIGS } from "../urls";
import * as router from "../router";

function convertNsToSecnod(value) {
  return Math.round(value / (1000 * 1000 * 1000));
}

function convertSecondToNs(value) {
  return value * (1000 * 1000 * 1000);
}

class UpdateConfig extends React.Component {
  state = {
    spinning: false,
    spinTips: "",
    data: null
  };
  async componentDidMount() {
    const { name } = this.props.match.params;
    this.setState({
      spinning: true,
      spinTips: "Loading..."
    });
    try {
      const { data } = await request.get(`${CONFIGS}/${name}`);
      if (data.cache) {
        data.cache.hitForPass = convertNsToSecnod(data.cache.hitForPass);
      }
      if (data.timeout) {
        const keys = Object.keys(data.timeout);
        keys.forEach(key => {
          data.timeout[key] = convertNsToSecnod(data.timeout[key]);
        });
      }
      this.setState({
        data
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        spinning: false
      });
    }
  }
  async handleSubmit(e) {
    e.preventDefault();
    const { name } = this.props.match.params;
    const { spinning, data } = this.state;
    if (spinning) {
      return;
    }
    this.setState({
      spinning: true,
      spinTips: "Updating..."
    });
    // TODO 做转换以及类型转换
    try {
      const updateData = Object.assign({}, data);
      const { cache, timeout } = updateData;
      if (cache && cache.hitForPass) {
        cache.hitForPass = convertSecondToNs(cache.hitForPass);
      }
      if (timeout) {
        const keys = Object.keys(timeout);
        keys.forEach(key => {
          timeout[key] = convertSecondToNs(timeout[key]);
        });
      }
      if (data.compress.filter) {
        // 测试是否正则
        new RegExp(data.compress.filter);
      }
      await request.patch(`${CONFIGS}/${name}`, updateData);
      message.info("update config successful");
      router.back();
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        spinning: false
      });
    }
  }
  createInputOnChagne(name) {
    const arr = name.split(".");
    return e => {
      const { target } = e;
      const { data } = this.state;
      let v = target.value.trim();
      // 针对相应的key转换为数字取 valueAsNumber
      if (!Number.isNaN(target.valueAsNumber)) {
        v = target.valueAsNumber;
      }
      const key = arr[0];
      let currentUpdate = data[key];
      if (arr.length === 1) {
        currentUpdate = v;
      } else {
        currentUpdate[arr[1]] = v;
      }
      const updateData = {};
      updateData[key] = currentUpdate;
      this.setState({
        data: Object.assign(data, updateData)
      });
    };
  }
  renderAdminConfig() {
    const { data } = this.state;
    const adminConfig = data.admin || {};
    return (
      <Card size="small" title="Admin Config" className="config">
        <Form.Item label="Prefix">
          <Input
            defaultValue={adminConfig.prefix}
            type="text"
            placeholder="Input the path prefix of admin page, e.g.: /pike"
            onChange={this.createInputOnChagne("admin.prefix")}
          />
        </Form.Item>
        <Form.Item label="User">
          <Input
            defaultValue={adminConfig.user}
            type="text"
            placeholder="Input the user of admin"
            onChange={this.createInputOnChagne("admin.user")}
          />
        </Form.Item>
        <Form.Item label="Password">
          <Input
            defaultValue={adminConfig.password}
            type="password"
            placeholder="Input the password of admin"
            onChange={this.createInputOnChagne("admin.password")}
          />
        </Form.Item>
      </Card>
    );
  }
  renderCacheConfig() {
    const { data } = this.state;
    return (
      <Card size="small" title="Cache Config" className="config">
        <Form.Item label="Zone">
          <Input
            defaultValue={data.cache.zone}
            type="number"
            placeholder="Input the max zone of cache"
            onChange={this.createInputOnChagne("cache.zone")}
          />
        </Form.Item>
        <Form.Item label="Size">
          <Input
            defaultValue={data.cache.size}
            type="number"
            placeholder="Input the max size of zone"
            onChange={this.createInputOnChagne("cache.size")}
          />
        </Form.Item>
        <Form.Item label="Hit For Pass">
          <Input
            defaultValue={data.cache.hitForPass}
            type="number"
            addonAfter="seconds"
            placeholder="Input the expire time of hit for pass"
            onChange={this.createInputOnChagne("cache.hitForPass")}
          />
        </Form.Item>
      </Card>
    );
  }
  renderCompressConfig() {
    const { data } = this.state;
    return (
      <Card size="small" title="Compress Config" className="config">
        <Form.Item label="Min Length">
          <Input
            addonAfter="bytes"
            type="number"
            defaultValue={data.compress.minLength}
            placeholder="Input the min lenght to compress"
            onChange={this.createInputOnChagne("compress.minLength")}
          />
        </Form.Item>
        <Form.Item label="Text Filter">
          <Input
            defaultValue={data.compress.filter}
            addonBefore="RegExp"
            placeholder="Input the text filter regexp"
            type="text"
            onChange={this.createInputOnChagne("compress.filter")}
          />
        </Form.Item>
      </Card>
    );
  }
  renderTimeoutConfig() {
    const { data } = this.state;
    const timeouts = [
      {
        name: "idleConn"
      },
      {
        name: "expectContinue"
      },
      {
        name: "responseHeader"
      },
      {
        name: "connect"
      },
      {
        name: "tlsHandshake"
      }
    ];
    const arr = timeouts.map(item => {
      const { name } = item;
      const value = data.timeout[name];
      const label = name[0].toUpperCase() + name.substring(1);
      return (
        <Form.Item key={name} label={label}>
          <Input
            defaultValue={value}
            type="number"
            addonAfter="seconds"
            placeholder={`Input the timeout for ${name}`}
            onChange={this.createInputOnChagne(`timeout.${name}`)}
          />
        </Form.Item>
      );
    });
    return (
      <Card size="small" title="Timeout Config" className="config">
        {arr}
      </Card>
    );
  }
  renderHeaderConfig({ name, title, placeholder }) {
    const { data } = this.state;
    return (
      <Card size="small" title={title} className="config">
        <Select
          mode="tags"
          placeholder={placeholder}
          defaultValue={data[name]}
          onChange={value => {
            const updateData = {};
            updateData[name] = value;
            this.setState({
              data: Object.assign(this.state.data, updateData)
            });
          }}
        />
      </Card>
    );
  }
  renderForm() {
    const { data } = this.state;
    if (!data) {
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
    return (
      <Form {...formItemLayout} onSubmit={this.handleSubmit.bind(this)}>
        {this.renderAdminConfig()}
        <Card size="small" title="Perfornamce Config" className="config">
          <Form.Item label="Concurrency">
            <Input
              defaultValue={data.concurrency}
              type="number"
              placeholder="Input max limit concurrency"
              onChange={this.createInputOnChagne("concurrency")}
            />
          </Form.Item>
          <Form.Item label="ServerTiming">
            <Switch
              defaultChecked={data.enableServerTiming}
              onChange={checked => {
                this.setState({
                  data: Object.assign(data, {
                    enableServerTiming: checked
                  })
                });
              }}
            />
          </Form.Item>
        </Card>
        {this.renderHeaderConfig({
          name: "responseHeader",
          title: "Response Header Config",
          placeholder:
            "Input the header will be add to response, e.g.: key:value"
        })}
        {this.renderHeaderConfig({
          name: "requestHeader",
          title: "Request Header Config",
          placeholder:
            "Input the header will be add to request, e.g.: key:value"
        })}
        {this.renderTimeoutConfig()}
        {this.renderCacheConfig()}
        {this.renderCompressConfig()}
        <Button type="primary" htmlType="submit" className="submit">
          Update
        </Button>
      </Form>
    );
  }
  render() {
    const { spinning, spinTips } = this.state;
    return (
      <div className="UpdateConfig">
        <Spin spinning={spinning} tip={spinTips}>
          {this.renderForm()}
        </Spin>
      </div>
    );
  }
}

export default UpdateConfig;
