import React from "react";
import PropTypes from "prop-types";
import _ from "lodash";
import {
  Form,
  Input,
  Spin,
  Button,
  Icon,
  Select,
  Switch,
  Row,
  Col
} from "antd";

import "./exform.sass";
import i18n from "../../i18n";

const { TextArea } = Input;
const { Option } = Select;

class KeyValueListInput extends React.Component {
  state = {
    keyValueList: null
  };
  constructor(props) {
    super(props);
    const value = props.value || [];
    const keyValueList = [];
    _.forEach(value, item => {
      const arr = item.split(":");
      keyValueList.push({
        key: arr[0],
        value: arr[1]
      });
    });
    if (keyValueList.length === 0) {
      keyValueList.push(null);
    }
    this.state.keyValueList = keyValueList;
  }
  handleChange(index, data) {
    const keyValueList = this.state.keyValueList.slice(0);
    if (!keyValueList[index]) {
      keyValueList[index] = {};
    }
    _.extend(keyValueList[index], data);
    this.setState({
      keyValueList
    });

    const { onChange } = this.props;
    if (onChange) {
      const values = [];
      _.forEach(keyValueList, item => {
        if (item.key && item.value) {
          values.push(`${item.key}:${item.value}`);
        }
      });
      onChange(values);
    }
  }
  renderKeyValue(item, index) {
    const placeholder = this.props.placeholder || [];
    return (
      <div key={`key-value-${index}`}>
        <Row gutter={8}>
          <Col span={11}>
            <Input
              defaultValue={item && item.key}
              placeholder={placeholder[0]}
              type="text"
              allowClear
              onChange={e => {
                this.handleChange(index, {
                  key: e.target.value
                });
              }}
            />
          </Col>
          <Col
            style={{
              textAlign: "center"
            }}
            span={1}
          >
            :
          </Col>
          <Col span={12}>
            <Input
              defaultValue={item && item.value}
              placeholder={placeholder[1]}
              type="text"
              allowClear
              onChange={e => {
                this.handleChange(index, {
                  value: e.target.value
                });
              }}
            />
          </Col>
        </Row>
      </div>
    );
  }
  render() {
    const { keyValueList } = this.state;
    const list = _.map(keyValueList, (item, index) => {
      return this.renderKeyValue(item, index);
    });
    return (
      <div>
        {list}
        <Button
          onClick={() => {
            const keyValueList = this.state.keyValueList.slice(0);
            keyValueList.push({});
            this.setState({
              keyValueList
            });
          }}
        >
          {i18n("common.add").toUpperCase()}
        </Button>
      </div>
    );
  }
}

class TextListInput extends React.Component {
  state = {
    textList: null
  };
  constructor(props) {
    super(props);
    const value = props.value || [];
    if (value.length === 0) {
      value.push("");
    }
    this.state.textList = value;
  }
  handleChange(index, value) {
    const textList = this.state.textList.slice(0);
    textList[index] = value;
    const { onChange } = this.props;
    if (onChange) {
      onChange(_.filter(textList, item => !!item));
    }
  }
  render() {
    const { textList } = this.state;
    const list = _.map(textList, (item, index) => {
      return (
        <Input
          defaultValue={item}
          key={`text-list-${index}`}
          type="text"
          placeholder={this.props.placeholder}
          allowClear
          onChange={e => {
            this.handleChange(index, e.target.value);
          }}
        />
      );
    });
    return (
      <div>
        {list}
        <Button
          onClick={() => {
            const textList = this.state.textList.slice(0);
            textList.push("");
            this.setState({
              textList
            });
          }}
        >
          {i18n("common.add").toUpperCase()}
        </Button>
      </div>
    );
  }
}

class UpstreamServersInput extends React.Component {
  state = {
    upstreamServers: null
  };
  constructor(props) {
    super(props);
    const value = props.value || [];
    if (value.length === 0) {
      value.push(null);
    }
    this.state.upstreamServers = value;
  }
  handleChange(index, value) {
    const servers = this.state.upstreamServers.slice(0);
    servers[index] = _.extend(servers[index], value);
    this.setState({
      upstreamServers: servers
    });
    const { onChange } = this.props;
    if (onChange) {
      onChange(_.filter(servers, item => !!item.addr));
    }
  }
  renderBackend(backend, index) {
    return (
      <div key={`backend-${index}`}>
        <Row gutter={8}>
          <Col span={20}>
            <Input
              onChange={e => {
                this.handleChange(index, {
                  addr: e.target.value
                });
              }}
              defaultValue={backend && backend.addr}
              type="text"
              placeholder={this.props.placeholder}
              allowClear
            />
          </Col>
          <Col span={3}>
            <Switch
              onChange={checked => {
                this.handleChange(index, {
                  backup: checked
                });
              }}
              checkedChildren="backup"
              unCheckedChildren="backup"
              defaultChecked={backend && backend.backup}
            />
          </Col>
        </Row>
      </div>
    );
  }
  render() {
    const { upstreamServers } = this.state;
    const servers = _.map(upstreamServers, (item, index) => {
      return this.renderBackend(item, index);
    });
    return (
      <div className="upstreamServers">
        {servers}
        <Button
          onClick={() => {
            const servers = this.state.upstreamServers.slice(0);
            servers.push({});
            this.setState({
              upstreamServers: servers
            });
          }}
        >
          {i18n("common.add").toUpperCase()}
        </Button>
      </div>
    );
  }
}

class ExForm extends React.Component {
  state = {
    spinning: false
  };
  handleSubmit = e => {
    e.preventDefault();
    const { fields, onSubmit } = this.props;
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (err) {
        return;
      }
      _.forEach(fields, item => {
        const { key, type } = item;
        if (type === "number" && values[key]) {
          values[key] = Number(values[key]);
        }
      });
      const done = () => {
        this.setState({
          spinning: false
        });
      };
      this.setState({
        spinning: true
      });
      onSubmit(values, done);
    });
  };
  render() {
    const { spinning } = this.state;
    const { onBack, form } = this.props;
    const { getFieldDecorator } = form;

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

    const { fields, title, description } = this.props;

    const originalData = this.props.originalData || {};

    const items = _.map(fields, item => {
      let decorator = null;
      let layout = null;
      const { key, rules } = item;
      const inputProps = {};
      if (key === "name" && originalData[key]) {
        inputProps.disabled = true;
      }
      const decoratorOpts = {
        rules,
        initialValue: originalData[key] || item.defaultValue
      };
      switch (item.type) {
        case "textarea":
          decorator = getFieldDecorator(
            key,
            decoratorOpts
          )(<TextArea rows={5} placeholder={item.placeholder || ""} />);
          break;
        case "select":
          const opts = _.map(item.options, item => {
            return (
              <Option key={item} value={item}>
                {item}
              </Option>
            );
          });
          decorator = getFieldDecorator(
            key,
            decoratorOpts
          )(<Select placeholder={item.placeholder || ""}>{opts}</Select>);
          break;
        case "upstreamServers":
          decorator = getFieldDecorator(
            key,
            decoratorOpts
          )(<UpstreamServersInput placeholder={item.placeholder || ""} />);
          break;
        case "textList":
          decorator = getFieldDecorator(
            key,
            decoratorOpts
          )(<TextListInput placeholder={item.placeholder} />);
          break;
        case "keyValueList":
          decorator = getFieldDecorator(
            key,
            decoratorOpts
          )(<KeyValueListInput placeholder={item.placeholder} />);
          break;
        default:
          decorator = getFieldDecorator(
            key,
            decoratorOpts
          )(
            <Input
              {...inputProps}
              placeholder={item.placeholder || ""}
              type={item.type || "text"}
            />
          );
      }
      return (
        <Form.Item label={item.label} key={item.key} {...layout}>
          {decorator}
        </Form.Item>
      );
    });
    items.push(
      <Form.Item key="submit" {...tailFormItemLayout}>
        <Button type="primary" htmlType="submit">
          {i18n("common.submit")}
        </Button>
      </Form.Item>
    );

    return (
      <div className="ExForm">
        <Spin spinning={spinning}>
          <h3>
            {onBack && (
              <a
                href="/back"
                className="back"
                onClick={e => {
                  e.preventDefault();
                  onBack();
                }}
              >
                <Icon type="left" />
                {i18n("common.back")}
              </a>
            )}
            {title}
          </h3>
          {description && <p>{description}</p>}
          <Form {...formItemLayout} onSubmit={this.handleSubmit}>
            {items}
          </Form>
        </Spin>
      </div>
    );
  }
}

const WrappedExForm = Form.create({ name: "exform" })(ExForm);

WrappedExForm.propTypes = {
  originalData: PropTypes.object,
  title: PropTypes.string.isRequired,
  description: PropTypes.string,
  fields: PropTypes.any.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onBack: PropTypes.func
};

export default WrappedExForm;
