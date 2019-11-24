import React from "react";
import PropTypes from "prop-types";
import _ from "lodash";
import { Form, Input, Spin, Button, Icon } from "antd";

import "./exform.sass";
import i18n from "../../i18n";

const { TextArea } = Input;

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
