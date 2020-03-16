import React from "react";
import { message } from "antd";
import axios from "axios";

import Configs from "../configs";
import { CONFIGS } from "../../urls";
import { getLocationI18n } from "../../i18n";

const category = "locations";
const renderList = row => {
  if (!row) {
    return;
  }
  const items = row.map(item => {
    return <li key={item}>{item}</li>;
  });
  return <ul>{items}</ul>;
};
const columns = [
  {
    title: getLocationI18n("name"),
    dataIndex: "name"
  },
  {
    title: getLocationI18n("upstream"),
    dataIndex: "upstream"
  },
  {
    title: getLocationI18n("hosts"),
    dataIndex: "hosts",
    render: renderList
  },
  {
    title: getLocationI18n("prefixs"),
    dataIndex: "prefixs",
    render: renderList
  },
  {
    title: getLocationI18n("rewrites"),
    dataIndex: "rewrites",
    render: renderList
  },
  {
    title: getLocationI18n("reqHeader"),
    dataIndex: "requestHeader",
    render: renderList
  },
  {
    title: getLocationI18n("resHeader"),
    dataIndex: "responseHeader",
    render: renderList
  }
];
const fields = [
  {
    label: getLocationI18n("name"),
    key: "name",
    placeholder: getLocationI18n("namePlaceHolder"),
    rules: [
      {
        required: true,
        message: getLocationI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getLocationI18n("upstream"),
    key: "upstream",
    placeholder: getLocationI18n("upstreamPlaceHolder"),
    rules: [
      {
        required: true,
        message: getLocationI18n("upstreamRequireMessage")
      }
    ],
    type: "select"
  },
  {
    label: getLocationI18n("hosts"),
    key: "hosts",
    placeholder: getLocationI18n("hostsPlaceHolder"),
    type: "textList"
  },
  {
    label: getLocationI18n("prefixs"),
    key: "prefixs",
    placeholder: getLocationI18n("prefixsPlaceHolder"),
    type: "textList"
  },
  {
    label: getLocationI18n("rewrites"),
    key: "rewrites",
    placeholder: [
      getLocationI18n("rewriteOriginalPlaceHolder"),
      getLocationI18n("rewriteNewPlaceHolder")
    ],
    type: "keyValueList"
  },
  {
    label: getLocationI18n("reqHeader"),
    key: "requestHeader",
    placeholder: [
      getLocationI18n("headerNamePlaceHolder"),
      getLocationI18n("headerValuePlaceHolder")
    ],
    type: "keyValueList"
  },
  {
    label: getLocationI18n("resHeader"),
    key: "responseHeader",
    placeholder: [
      getLocationI18n("headerNamePlaceHolder"),
      getLocationI18n("headerValuePlaceHolder")
    ],
    type: "keyValueList"
  }
];

class Locations extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: getLocationI18n("createUpdateTitle"),
      description: getLocationI18n("createUpdateDescription"),
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
      const upstreams = data.upstreams.map(item => item.name);
      fields.forEach(item => {
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
