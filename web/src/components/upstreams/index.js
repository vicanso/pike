import React from "react";

import Configs from "../configs";
import { getCommonI18n, getUpstreamI18n } from "../../i18n";

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
          backupTips = <span className="backupTips">backup</span>;
        }
        return (
          <li key={item.addr}>
            {item.addr}
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
      category,
      columns,
      fields
    });
  }
}

export default Upstreams;
