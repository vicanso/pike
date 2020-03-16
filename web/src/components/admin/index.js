import React from "react";
import { Switch } from "antd";

import Configs from "../configs";

import { getAdminI18n, getCommonI18n } from "../../i18n";

const category = "admin";

const columns = [
  {
    title: getAdminI18n("user"),
    dataIndex: "user"
  },
  {
    title: getAdminI18n("password"),
    dataIndex: "password",
    render: row => {
      if (row) {
        return "***";
      }
      return "";
    }
  },
  {
    title: getAdminI18n("prefix"),
    dataIndex: "prefix"
  },
  {
    title: getAdminI18n("enabledInternetAccess"),
    dataIndex: "enabledInternetAccess",
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
    label: getAdminI18n("user"),
    key: "user",
    placeholder: getAdminI18n("userPlaceHolder")
  },
  {
    label: getAdminI18n("password"),
    key: "password",
    placeholder: getAdminI18n("passwordPlaceHolder")
  },
  {
    label: getAdminI18n("prefix"),
    key: "prefix",
    placeholder: getAdminI18n("prefixPlaceHolder"),
    rules: [
      {
        required: true,
        message: getAdminI18n("prefixRequireMessage")
      }
    ]
  },
  {
    label: getAdminI18n("enabledInternetAccess"),
    key: "enabledInternetAccess",
    type: "switch"
  },
  {
    label: getCommonI18n("description"),
    key: "description",
    type: "textarea",
    placeholder: getCommonI18n("descriptionPlaceholder")
  }
];

class Admin extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      disabledDelete: true,
      title: getAdminI18n("createUpdateTitle"),
      description: getAdminI18n("createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Admin;
