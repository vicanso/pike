import React from "react";
import _ from "lodash";
import { message } from "antd";
import axios from "axios";

import i18n from "../../i18n";
import Configs from "../configs";
import { CONFIGS } from "../../urls";

const category = "locations";
const renderList = row => {
  const items = _.map(row, item => {
    return <li key={item}>{item}</li>;
  });
  return <ul>{items}</ul>;
};
const columns = [
  {
    title: i18n("location.name"),
    dataIndex: "name"
  },
  {
    title: i18n("location.upstream"),
    dataIndex: "upstream"
  },
  {
    title: i18n("location.hosts"),
    dataIndex: "hosts",
    render: renderList
  },
  {
    title: i18n("location.prefixs"),
    dataIndex: "prefixs",
    render: renderList
  },
  {
    title: i18n("location.rewrites"),
    dataIndex: "rewrites",
    render: renderList
  },
  {
    title: i18n("location.reqHeader"),
    dataIndex: "requestHeader",
    render: renderList
  },
  {
    title: i18n("location.resHeader"),
    dataIndex: "responseHeader",
    render: renderList
  }
];
const fields = [
  {
    label: i18n("location.name"),
    key: "name",
    placeholder: i18n("location.namePlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("location.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("location.upstream"),
    key: "upstream",
    placeholder: i18n("location.upstreamPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("location.upstreamRequireMessage")
      }
    ],
    type: "select"
  },
  {
    label: i18n("location.hosts"),
    key: "hosts",
    placeholder: i18n("location.hostsPlaceHolder"),
    type: "textList"
  },
  {
    label: i18n("location.prefixs"),
    key: "prefixs",
    placeholder: i18n("location.prefixsPlaceHolder"),
    type: "textList"
  },
  {
    label: i18n("location.rewrites"),
    key: "rewrites",
    placeholder: [
      i18n("location.rewriteOriginalPlaceHolder"),
      i18n("location.rewriteNewPlaceHolder")
    ],
    type: "keyValueList"
  },
  {
    label: i18n("location.reqHeader"),
    key: "requestHeader",
    placeholder: [
      i18n("location.headerNamePlaceHolder"),
      i18n("location.headerValuePlaceHolder")
    ],
    type: "keyValueList"
  },
  {
    label: i18n("location.resHeader"),
    key: "responseHeader",
    placeholder: [
      i18n("location.headerNamePlaceHolder"),
      i18n("location.headerValuePlaceHolder")
    ],
    type: "keyValueList"
  }
];

class Locations extends Configs {
  constructor(props) {
    super(props);
    _.assignIn(this.state, {
      title: i18n("location.createUpdateTitle"),
      description: i18n("location.createUpdateDescription"),
      category,
      columns
    });
  }
  async componentDidMount() {
    super.componentDidMount();
    try {
      const { data } = await axios.get(
        CONFIGS.replace(":category", "upstreams")
      );
      const upstreams = _.map(data.upstreams, item => item.name);
      _.forEach(fields, item => {
        if (item.key === "upstream") {
          item.options = upstreams;
        }
      });
      this.setState({
        fields
      });
    } catch (err) {
      message.error(err.message);
    } finally {
    }
  }
}

export default Locations;
