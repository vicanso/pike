import React from "react";
import { message, Switch } from "antd";
import axios from "axios";

import i18n from "../../i18n";
import Configs from "../configs";
import { CONFIGS } from "../../urls";
import { numberToDuration } from "../../util";

const category = "servers";

const renderDuration = row => {
  return <span>{numberToDuration(row)}</span>;
};

const columns = [
  {
    title: i18n("server.name"),
    dataIndex: "name"
  },
  {
    title: i18n("server.addr"),
    dataIndex: "addr"
  },
  {
    title: i18n("server.cache"),
    dataIndex: "cache"
  },
  {
    title: i18n("server.compress"),
    dataIndex: "compress"
  },
  {
    title: i18n("server.locations"),
    dataIndex: "locations",
    render: row => {
      const locations = row.map(item => {
        return <li key={item}>{item}</li>;
      });
      return <ul>{locations}</ul>;
    }
  },
  {
    title: i18n("server.etag"),
    dataIndex: "eTag",
    render: row => {
      return <Switch disabled={true} defaultChecked={row} />;
    }
  },
  {
    title: i18n("server.concurrency"),
    dataIndex: "concurrency"
  },
  {
    title: i18n("server.readTimeout"),
    dataIndex: "readTimeout",
    render: renderDuration
  },
  {
    title: i18n("server.writeTimeout"),
    dataIndex: "writeTimeout",
    render: renderDuration
  },
  {
    title: i18n("server.idleTimeout"),
    dataIndex: "idleTimeout",
    render: renderDuration
  },
  {
    title: i18n("server.maxHeaderBytes"),
    dataIndex: "maxHeaderBytes"
  }
];
const fields = [
  {
    label: i18n("server.name"),
    key: "name",
    placeholder: i18n("server.namePlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("server.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("server.addr"),
    key: "addr",
    placeholder: i18n("server.addrPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("server.addrRequireMessage")
      }
    ]
  },
  {
    label: i18n("server.cache"),
    key: "cache",
    placeholder: i18n("server.cachePlaceHolder"),
    type: "select",
    rules: [
      {
        required: true,
        message: i18n("server.cacheRequireMessage")
      }
    ]
  },
  {
    label: i18n("server.compress"),
    key: "compress",
    placeholder: i18n("server.compressPlaceHolder"),
    type: "select",
    rules: [
      {
        required: true,
        message: i18n("server.compressRequireMessage")
      }
    ]
  },
  {
    label: i18n("server.locations"),
    key: "locations",
    placeholder: i18n("server.locationsPlaceHolder"),
    type: "select",
    mode: "multiple",
    rules: [
      {
        required: true,
        message: i18n("server.locationsRequireMesage")
      }
    ]
  },
  {
    label: i18n("server.etag"),
    key: "eTag",
    type: "switch"
  },
  {
    label: i18n("server.concurrency"),
    key: "concurrency",
    type: "number",
    placeholder: i18n("server.concurrencyPlaceHolder")
  },
  {
    label: i18n("server.readTimeout"),
    key: "readTimeout",
    type: "duration",
    placeholder: i18n("server.readTimeoutPlaceHolder")
  },
  {
    label: i18n("server.writeTimeout"),
    key: "writeTimeout",
    type: "duration",
    placeholder: i18n("server.writeTimeoutPlaceHolder")
  },
  {
    label: i18n("server.idleTimeout"),
    key: "idleTimeout",
    type: "duration",
    placeholder: i18n("server.idleTimeoutPlaceHolder")
  },
  {
    label: i18n("server.maxHeaderBytes"),
    key: "maxHeaderBytes",
    type: "number",
    placeholder: i18n("server.maxHeaderBytesPlaceHolder")
  }
];

class Servers extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: i18n("server.createUpdateTitle"),
      description: i18n("server.createUpdateDescription"),
      columns,
      category
    });
  }
  async componentDidMount() {
    super.componentDidMount();
    try {
      const cat = ["caches", "compresses", "locations"].join(",");
      const { data } = await axios.get(CONFIGS.replace(":category", cat));
      const caches = data.caches.map(item => item.name);
      const compresses = data.compresses.map(item => item.name);
      const locations = data.locations.map(item => item.name);
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
