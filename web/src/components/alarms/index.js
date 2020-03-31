import React from "react";
import Configs from "../configs";
import { getAlarmI18n } from "../../i18n";

const category = "alarms";

const columns = [
  {
    title: getAlarmI18n("name"),
    dataIndex: "name"
  },
  {
    title: getAlarmI18n("uri"),
    dataIndex: "uri"
  },
  {
    title: getAlarmI18n("template"),
    dataIndex: "template",
    render: row => {
      if (!row) {
        return;
      }
      return <pre>{row}</pre>;
    }
  }
];

const fields = [
  {
    label: getAlarmI18n("name"),
    key: "name",
    placeholder: getAlarmI18n("namePlaceHolder"),
    rules: [
      {
        required: true,
        message: getAlarmI18n("nameRequireMessage")
      }
    ],
    options: ["upstream"],
    type: "select"
  },
  {
    label: getAlarmI18n("uri"),
    key: "uri",
    placeholder: getAlarmI18n("uriPlaceHolder"),
    rules: [
      {
        required: true,
        message: getAlarmI18n("uriRequireMessage")
      }
    ]
  },
  {
    label: getAlarmI18n("template"),
    key: "template",
    placeholder: getAlarmI18n("templatePlaceHolder"),
    rules: [
      {
        required: true,
        message: getAlarmI18n("templateRequireMessage")
      }
    ],
    type: "textarea"
  }
];

class Alarms extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: getAlarmI18n("createUpdateTitle"),
      description: getAlarmI18n("createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Alarms;
