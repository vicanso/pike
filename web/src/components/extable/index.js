import React from "react";
import PropTypes from "prop-types";
import { Table, Popconfirm, Icon, Spin, message } from "antd";

import i18n from "../../i18n";
import "./extable.sass";

class ExTable extends React.Component {
  state = {
    dataSource: null,
    submitting: false
  };
  async handleDelete(item) {
    this.setState({
      submitting: true
    });
    try {
      await this.props.onDelete(item);
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        submitting: false
      });
    }
  }
  render() {
    const { submitting } = this.state;
    const { columns, rowKey, dataSource, onDelete, onUpdate } = this.props;
    const cloneColumns = columns.slice(0);
    // 只有设置了更新或删除函数才添加功能操作列表
    if (onDelete || onUpdate) {
      cloneColumns.push({
        title: i18n("common.action"),
        render: row => {
          return (
            <div className="action">
              {onDelete && (
                <Popconfirm
                  title={i18n("common.deleteTips")}
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
                    {i18n("common.delete")}
                  </a>
                </Popconfirm>
              )}
              {onUpdate && (
                <a
                  href="/update"
                  onClick={e => {
                    e.preventDefault();
                    onUpdate(row);
                  }}
                >
                  <Icon type="edit" />
                  {i18n("common.update")}
                </a>
              )}
            </div>
          );
        }
      });
    }

    return (
      <div className="ExTable">
        <Spin spinning={submitting}>
          <Table
            rowKey={rowKey || "name"}
            className="ExTable"
            dataSource={dataSource}
            columns={cloneColumns}
          />
        </Spin>
      </div>
    );
  }
}

ExTable.propTypes = {
  columns: PropTypes.array.isRequired,
  dataSource: PropTypes.array,
  rowKey: PropTypes.string,
  onUpdate: PropTypes.func,
  onDelete: PropTypes.func
};

export default ExTable;
