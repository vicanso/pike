import React from "react";
import { message, Switch } from "antd";
import axios from "axios";

import Configs from "../configs";
import { CONFIGS } from "../../urls";
import { numberToDuration } from "../../util";
import { getServerI18n } from "../../i18n";
const category = "servers";

const renderDuration = row => {
  return <span>{numberToDuration(row)}</span>;
};
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
    title: getServerI18n("name"),
    dataIndex: "name",
    fixed: "left"
  },
  {
    title: getServerI18n("addr"),
    dataIndex: "addr"
  },
  {
    title: getServerI18n("cache"),
    dataIndex: "cache"
  },
  {
    title: getServerI18n("compress"),
    dataIndex: "compress"
  },
  {
    title: getServerI18n("locations"),
    dataIndex: "locations",
    render: renderList
  },
  {
    title: getServerI18n("certs"),
    dataIndex: "certs",
    render: renderList
  },
  {
    title: getServerI18n("http3"),
    dataIndex: "http3",
    render: row => {
      return <Switch disabled={true} defaultChecked={row} />;
    }
  },
  {
    title: getServerI18n("etag"),
    dataIndex: "eTag",
    render: row => {
      return <Switch disabled={true} defaultChecked={row} />;
    }
  },
  {
    title: getServerI18n("concurrency"),
    dataIndex: "concurrency"
  },
  {
    title: getServerI18n("readTimeout"),
    dataIndex: "readTimeout",
    render: renderDuration
  },
  {
    title: getServerI18n("writeTimeout"),
    dataIndex: "writeTimeout",
    render: renderDuration
  },
  {
    title: getServerI18n("idleTimeout"),
    dataIndex: "idleTimeout",
    render: renderDuration
  },
  {
    title: getServerI18n("maxHeaderBytes"),
    dataIndex: "maxHeaderBytes"
  }
];
const fields = [
  {
    label: getServerI18n("name"),
    key: "name",
    placeholder: getServerI18n("namePlaceHolder"),
    rules: [
      {
        required: true,
        message: getServerI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getServerI18n("addr"),
    key: "addr",
    placeholder: getServerI18n("addrPlaceHolder"),
    rules: [
      {
        required: true,
        message: getServerI18n("addrRequireMessage")
      }
    ]
  },
  {
    label: getServerI18n("cache"),
    key: "cache",
    placeholder: getServerI18n("cachePlaceHolder"),
    type: "select",
    rules: [
      {
        required: true,
        message: getServerI18n("cacheRequireMessage")
      }
    ]
  },
  {
    label: getServerI18n("compress"),
    key: "compress",
    placeholder: getServerI18n("compressPlaceHolder"),
    type: "select",
    rules: [
      {
        required: true,
        message: getServerI18n("compressRequireMessage")
      }
    ]
  },
  {
    label: getServerI18n("locations"),
    key: "locations",
    placeholder: getServerI18n("locationsPlaceHolder"),
    type: "select",
    mode: "multiple",
    rules: [
      {
        required: true,
        message: getServerI18n("locationsRequireMesage")
      }
    ]
  },
  {
    label: getServerI18n("certs"),
    key: "certs",
    placeholder: getServerI18n("certsPlaceHolder"),
    type: "select",
    mode: "multiple"
  },
  {
    label: getServerI18n("http3"),
    key: "http3",
    type: "switch"
  },
  {
    label: getServerI18n("etag"),
    key: "eTag",
    type: "switch"
  },
  {
    label: getServerI18n("concurrency"),
    key: "concurrency",
    type: "number",
    placeholder: getServerI18n("concurrencyPlaceHolder")
  },
  {
    label: getServerI18n("readTimeout"),
    key: "readTimeout",
    type: "duration",
    placeholder: getServerI18n("readTimeoutPlaceHolder")
  },
  {
    label: getServerI18n("writeTimeout"),
    key: "writeTimeout",
    type: "duration",
    placeholder: getServerI18n("writeTimeoutPlaceHolder")
  },
  {
    label: getServerI18n("idleTimeout"),
    key: "idleTimeout",
    type: "duration",
    placeholder: getServerI18n("idleTimeoutPlaceHolder")
  },
  {
    label: getServerI18n("maxHeaderBytes"),
    key: "maxHeaderBytes",
    type: "number",
    placeholder: getServerI18n("maxHeaderBytesPlaceHolder")
  }
];

class Servers extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      minWidth: 1800,
      title: getServerI18n("createUpdateTitle"),
      description: getServerI18n("createUpdateDescription"),
      columns,
      category
    });
  }
  async componentDidMount() {
    super.componentDidMount();
    try {
      const cat = ["caches", "compresses", "locations", "certs"].join(",");
      const { data } = await axios.get(CONFIGS.replace(":category", cat));
      const caches = data.caches.map(item => item.name);
      const compresses = data.compresses.map(item => item.name);
      const locations = data.locations.map(item => item.name);
      const certs = data.certs.map(item => item.name);
      fields.forEach(item => {
        switch (item.key) {
          case "cache":
            item.options = caches;
            break;
          case "compress":
            item.options = compresses;
            break;
          case "locations":
            item.options = locations;
            break;
          case "certs":
            item.options = certs;
            break;
          default:
            break;
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

export default Servers;
