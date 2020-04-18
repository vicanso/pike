import React from "react";
import axios from "axios";
import { Icon, Switch } from "antd";

import Configs from "../configs";
import { getCommonI18n, getUpstreamI18n } from "../../i18n";
import { UPSTREAMS } from "../../urls";

const category = "upstreams";
const columns = [
  {
    title: getUpstreamI18n("name"),
    dataIndex: "name"
  },
  {
    title: getUpstreamI18n("policy"),
    dataIndex: "policy"
  },
  {
    title: getUpstreamI18n("servers"),
    dataIndex: "servers",
    render: row => {
      const servers = row.map(item => {
        let backupTips = null;
        if (item.backup) {
          backupTips = <span className="tips">backup</span>;
        }
        let iconType = "";
        const style = {
          margin: "0 5px"
        };
        switch (item.status) {
          case "healthy":
            style.color = "#1890ff";
            iconType = "check-circle";
            break;
          case "sick":
            style.color = "#ff4d4f";
            iconType = "close-circle";
            break;
          default:
            iconType = "question-circle";
            break;
        }
        return (
          <li key={item.addr}>
            {item.addr}
            <Icon style={style} type={iconType} />
            {backupTips}
          </li>
        );
      });
      return <ul className="upstreamServers">{servers}</ul>;
    }
  },
  {
    title: getUpstreamI18n("healthCheck"),
    dataIndex: "healthCheck"
  },
  {
    title: getUpstreamI18n("h2c"),
    dataIndex: "enableH2C",
    render: row => {
      return <Switch disabled={true} defaultChecked={row} />;
    }
  },
  {
    title: getCommonI18n("description"),
    dataIndex: "description"
  }
];
const fields = [
  {
    label: getUpstreamI18n("name"),
    key: "name",
    placeholder: getUpstreamI18n("namePlaceHolder"),
    rules: [
      {
        required: true,
        message: getUpstreamI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getUpstreamI18n("servers"),
    key: "servers",
    type: "upstreamServers",
    placeholder: getUpstreamI18n("serverAddrPlaceHolder"),
    rules: [
      {
        required: true,
        message: getUpstreamI18n("serversRequireMessage")
      }
    ]
  },
  {
    label: getUpstreamI18n("policy"),
    key: "policy",
    placeholder: getUpstreamI18n("policyPlaceHolder"),
    type: "select",
    options: ["roundRobin", "first", "random", "leastconn"]
  },
  {
    label: getUpstreamI18n("healthCheck"),
    key: "healthCheck",
    placeholder: getUpstreamI18n("healthCheckPlaceHolder")
  },
  {
    label: getUpstreamI18n("h2c"),
    key: "enableH2C",
    type: "switch",
    title: getUpstreamI18n("h2cTitle"),
  },
  {
    label: getCommonI18n("description"),
    key: "description",
    type: "textarea",
    placeholder: getCommonI18n("descriptionPlaceholder")
  }
];

class Upstreams extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: getUpstreamI18n("createUpdateTitle"),
      description: getUpstreamI18n("createUpdateDescription"),
      handleConfigs: this.updateServerInfo,
      category,
      columns,
      fields
    });
  }
  async updateServerInfo(servers) {
    const { data } = await axios.get(UPSTREAMS);
    servers.forEach(item => {
      const upstreams = data[item.name];
      if (!upstreams) {
        return;
      }
      // 获取server中upstream的状态
      item.servers.forEach(server => {
        upstreams.forEach(upstream => {
          if (upstream.url === server.addr) {
            server.status = upstream.status;
          }
        });
      });
    });
  }
}

export default Upstreams;
