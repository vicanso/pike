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

class UpstreamServersInput extends React.Component {
  state = {
    upstreamServers: null
  };
  constructor(props) {
    super(props);
    const value = props.value || [];
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
      <li key={`backend-${index}`}>
        <Row gutter={8}>
          <Col span={19}>
            <Input
              onChange={e => {
                this.handleChange(index, {
                  addr: e.target.value
                });
              }}
              defaultValue={backend && backend.addr}
              type="text"
              placeholder={this.props.placeholder}
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
          <Col className="remove" span={2}>
            <a
              href="/remove-backend"
              onClick={e => {
                e.preventDefault();
                const arr = this.state.upstreamServers.slice(0);
                // 为了保证数组的顺序，直接置空，不删除
                arr[index] = {};
                this.setState({
                  upstreamServers: arr
                });
              }}
            >
              <Icon type="close-circle" />
            </a>
          </Col>
        </Row>
      </li>
    );
  }
  render() {
    const { upstreamServers } = this.state;
    const servers = _.map(upstreamServers, (item, index) => {
      return this.renderBackend(item, index);
    });
    return (
      <ul className="upstreamServers">
        {servers}
        {this.renderBackend(null, servers.length)}
      </ul>
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
      switch (item.type) {
        case "textarea":
          decorator = getFieldDecorator(key, {
            rules,
            initialValue: originalData[key]
          })(<TextArea rows={5} placeholder={item.placeholder || ""} />);
          break;
        case "select":
          const opts = _.map(item.options, item => {
            return (
              <Option key={item} value={item}>
                {item}
              </Option>
            );
          });
          decorator = getFieldDecorator(key, {
            initialValue: originalData[key]
          })(<Select placeholder={item.placeholder || ""}>{opts}</Select>);
          break;
        case "upstreamServers":
          decorator = getFieldDecorator(key, {
            initialValue: originalData[key]
          })(<UpstreamServersInput placeholder={item.placeholder || ""} />);
          break;
        default:
          decorator = getFieldDecorator(key, {
            rules,
            initialValue: originalData[key] || item.defaultValue
          })(
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
