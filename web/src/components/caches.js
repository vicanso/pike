import React from "react";
import { Input, message, Spin, Tabs, Table, Modal } from "antd";
import request from "axios";

import { CACHES } from "../urls";
import { getExpiredDesc } from "../util";

import "./caches.sass";

const TabPane = Tabs.TabPane;

const Fetch = 0;
const HitForPass = 2;

class Caches extends React.Component {
  state = {
    loading: false,
    confirmDelete: false,
    deleteCacheKey: "",
    caches: null
  };
  async handleDelete() {
    const { deleteCacheKey, caches } = this.state;
    this.setState({
      confirmDelete: false
    });
    const done = message.loading("Are you sure to delete the cache?", 0);
    try {
      const url = `${CACHES}?key=${deleteCacheKey}`;
      await request.delete(url);
      const result = [];
      caches.forEach(item => {
        if (item.key !== deleteCacheKey) {
          result.push(item);
        }
      });
      this.setState({
        caches: result
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      done();
    }
  }
  handleCancel() {
    this.setState({
      confirmDelete: false
    });
  }
  renderCaches() {
    const { loading, caches } = this.state;
    if (loading || !caches) {
      return;
    }
    const fechingList = [];
    const hitFroPassList = [];
    const cachedList = [];
    const now = Math.floor(Date.now() / 1000);
    caches.forEach(item => {
      const created = (item.expiredAt - item.maxAge) * 1000;
      const formatItem = {
        key: item.key,
        maxAge: item.maxAge,
        expired: item.expiredAt,
        expiredDesc: getExpiredDesc(item.expiredAt - now),
        created: created,
        createdDesc: new Date(created).toLocaleString()
      };
      switch (item.status) {
        case Fetch:
          fechingList.push(formatItem);
          break;
        case HitForPass:
          hitFroPassList.push(formatItem);
          break;
        default:
          cachedList.push(formatItem);
          break;
      }
    });
    const columns = [
      {
        title: "Key",
        dataIndex: "key",
        key: "key",
        sorter: (a, b) => {
          if (a.key > b.key) {
            return 1;
          }
          return -1;
        }
      },
      {
        title: "MaxAge",
        dataIndex: "maxAge",
        key: "maxAge",
        sorter: (a, b) => a.maxAge - b.maxAge
      },
      {
        title: "Expired",
        dataIndex: "expiredDesc",
        key: "expired",
        sorter: (a, b) => a.expired - b.expired
      },
      {
        title: "Created",
        dataIndex: "createdDesc",
        key: "created",
        sorter: (a, b) => a.created - b.created
      },
      {
        title: "Action",
        dataIndex: "",
        key: "x",
        render: row => {
          return (
            <a
              href="/delete"
              onClick={e => {
                e.preventDefault();
                this.setState({
                  deleteCacheKey: row.key,
                  confirmDelete: true
                });
              }}
            >
              Delete
            </a>
          );
        }
      }
    ];
    return (
      <div>
        <Modal
          title="Delete Cache"
          visible={this.state.confirmDelete}
          onOk={this.handleDelete.bind(this)}
          onCancel={this.handleCancel.bind(this)}
        >
          <p>Do you want to delete the cache: {this.state.deleteCacheKey}</p>
        </Modal>
        <Input addonBefore="keyword" size="large" allowClear />
        <Tabs defaultActiveKey="1" className="tabs">
          <TabPane tab={`Cache(${cachedList.length})`} key="1">
            <Table
              className="table"
              dataSource={cachedList}
              columns={columns}
            />
          </TabPane>
          <TabPane tab={`HitForPass(${hitFroPassList.length})`} key="2">
            <Table
              className="table"
              dataSource={hitFroPassList}
              columns={columns}
            />
          </TabPane>
          <TabPane tab={`Fetching(${fechingList.length})`} key="3">
            <Table
              className="table"
              dataSource={fechingList}
              columns={columns}
            />
          </TabPane>
        </Tabs>
      </div>
    );
  }
  render() {
    const { loading } = this.state;
    return (
      <div className="Caches">
        {loading && (
          <div
            style={{
              textAlign: "center",
              paddingTop: "50px"
            }}
          >
            <Spin tip="Loading..." />
          </div>
        )}
        {this.renderCaches()}
      </div>
    );
  }
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await request.get(CACHES);
      this.setState({
        caches: data.caches || []
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        loading: false
      });
    }
  }
}

export default Caches;
