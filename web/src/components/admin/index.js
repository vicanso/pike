import React from "react";
import _ from "lodash";
import { Switch } from "antd";

import i18n from "../../i18n";
import Configs from "../configs";

const category = "admin";

const columns = [
  {
    title: i18n("admin.user"),
    dataIndex: "user"
  },
  {
    title: i18n("admin.password"),
    dataIndex: "password"
  },
  {
    title: i18n("admin.prefix"),
    dataIndex: "prefix"
  },
  {
    title: i18n("admin.enabledInternetAccess"),
    dataIndex: "enabledInternetAccess",
    render: row => {
      return <Switch disabled={true} defaultChecked={row} />;
    }
  },
  {
    title: i18n("common.description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: i18n("admin.user"),
    key: "user",
    placeholder: i18n("admin.userPlaceHolder")
  },
  {
    label: i18n("admin.password"),
    key: "password",
    placeholder: i18n("admin.passwordPlaceHolder")
  },
  {
    label: i18n("admin.prefix"),
    key: "prefix",
    placeholder: i18n("admin.prefixPlaceHolder")
  },
  {
    label: i18n("admin.enabledInternetAccess"),
    key: "enabledInternetAccess",
    type: "switch"
  },
  {
    label: i18n("common.description"),
    key: "description",
    type: "textarea",
    placeholder: i18n("common.descriptionPlaceholder")
  }
];

class Admin extends Configs {
  constructor(props) {
    super(props);
    _.assignIn(this.state, {
      title: i18n("admin.createUpdateTitle"),
      description: i18n("admin.createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Admin;
