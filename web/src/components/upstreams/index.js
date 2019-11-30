import React from "react";
import _ from "lodash";

import i18n from "../../i18n";
import Configs from "../configs";

const category = "upstreams";
const columns = [
  {
    title: i18n("upstream.name"),
    dataIndex: "name"
  },
  {
    title: i18n("upstream.policy"),
    dataIndex: "policy"
  },
  {
    title: i18n("upstream.servers"),
    dataIndex: "servers",
    render: row => {
      const servers = _.map(row, item => {
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
    title: i18n("upstream.healthCheck"),
    dataIndex: "healthCheck"
  },
  {
    title: i18n("common.description"),
    dataIndex: "description"
  }
];
const fields = [
  {
    label: i18n("upstream.name"),
    key: "name",
    placeholder: i18n("upstream.namePlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("upstream.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("upstream.servers"),
    key: "servers",
    type: "upstreamServers",
    placeholder: i18n("upstream.serverAddrPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("upstream.serversRequireMessage")
      }
    ]
  },
  {
    label: i18n("upstream.policy"),
    key: "policy",
    placeholder: i18n("upstream.policyPlaceHolder"),
    type: "select",
    options: ["roundRobin", "first", "random", "leastconn"]
  },
  {
    label: i18n("upstream.healthCheck"),
    key: "healthCheck",
    placeholder: i18n("upstream.healthCheckPlaceHolder")
  },
  {
    label: i18n("common.description"),
    key: "description",
    type: "textarea",
    placeholder: i18n("common.descriptionPlaceholder")
  }
];

class Upstreams extends Configs {
  constructor(props) {
    super(props);
    _.assignIn(this.state, {
      title: i18n("upstream.createUpdateTitle"),
      description: i18n("upstream.createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Upstreams;
