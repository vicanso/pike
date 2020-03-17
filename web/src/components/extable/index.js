import React from "react";
import PropTypes from "prop-types";
import { Table, Popconfirm, Icon, Spin, message } from "antd";

import { getCommonI18n } from "../../i18n";
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
    const {
      columns,
      rowKey,
      dataSource,
      onDelete,
      onUpdate,
      minWidth
    } = this.props;
    const cloneColumns = columns.slice(0);

    const scroll = {};
    let actionColumnFixed = "";
    if (minWidth && window && minWidth > window.innerWidth) {
      scroll.x = minWidth;
      actionColumnFixed = "right";
    }
    // 只有设置了更新或删除函数才添加功能操作列表
    if (onDelete || onUpdate) {
      cloneColumns.push({
        title: getCommonI18n("action"),
        width: 200,
        fixed: actionColumnFixed,
        render: row => {
          return (
            <div className="action">
              {onDelete && (
                <Popconfirm
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
                  {getCommonI18n("update")}
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
            scroll={scroll}
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
  onDelete: PropTypes.func,
  minWidth: PropTypes.number
};

export default ExTable;
