import React from "react";
import { Switch } from "antd";

import Configs from "../configs";

import { getInfluxdbI18n, getCommonI18n } from "../../i18n";

const category = "influxdb";

const columns = [
  {
    title: getInfluxdbI18n("uri"),
    dataIndex: "uri"
  },
  {
    title: getInfluxdbI18n("bucket"),
    dataIndex: "bucket"
  },
  {
    title: getInfluxdbI18n("org"),
    dataIndex: "org"
  },
  {
    title: getInfluxdbI18n("token"),
    dataIndex: "token",
    render: row => {
      if (!row) {
        return;
      }
      return `${row.substring(0, 8)}...`;
    }
  },
  {
    title: getInfluxdbI18n("batchSize"),
    dataIndex: "batchSize"
  },
  {
    title: getInfluxdbI18n("flushInterval"),
    dataIndex: "flushInterval"
  },
  {
    title: getInfluxdbI18n("enabled"),
    dataIndex: "enabled",
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
    label: getInfluxdbI18n("uri"),
    key: "uri",
    placeholder: getInfluxdbI18n("uriPlaceHolder"),
    rules: [
      {
        required: true
      }
    ]
  },
  {
    label: getInfluxdbI18n("bucket"),
    key: "bucket",
    placeholder: getInfluxdbI18n("bucketPlaceHolder"),
    rules: [
      {
        required: true
      }
    ]
  },
  {
    label: getInfluxdbI18n("org"),
    key: "org",
    placeholder: getInfluxdbI18n("orgPlaceHolder"),
    rules: [
      {
        required: true
      }
    ]
  },
  {
    label: getInfluxdbI18n("token"),
    key: "token",
    placeholder: getInfluxdbI18n("tokenPlaceHolder"),
    rules: [
      {
        required: true
      }
    ]
  },
  {
    label: getInfluxdbI18n("batchSize"),
    key: "batchSize",
    type: "number",
    placeholder: getInfluxdbI18n("batchSizePlaceHolder"),
    rules: [
      {
        required: true
      }
    ]
  },
  {
    label: getInfluxdbI18n("flushInterval"),
    key: "flushInterval",
    type: "number",
    placeholder: getInfluxdbI18n("flushIntervalPlaceHolder")
  },
  {
    label: getInfluxdbI18n("enabled"),
    key: "enabled",
    type: "switch"
  },
  {
    label: getCommonI18n("description"),
    key: "description",
    type: "textarea",
    placeholder: getCommonI18n("descriptionPlaceholder")
  }
];

class Influxdb extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      disabledDelete: true,
      single: true,
      title: getInfluxdbI18n("title"),
      description: getInfluxdbI18n("description"),
      category,
      columns,
      fields
    });
  }
}

export default Influxdb;
