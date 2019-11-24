import _ from "lodash";

import i18n from "../../i18n";
import Configs from "../configs";

const category = "compresses";

const columns = [
  {
    title: i18n("compresses.name"),
    dataIndex: "name"
  },
  {
    title: i18n("compresses.level"),
    dataIndex: "level",
    sorter: (a, b) => a.level - b.level
  },
  {
    title: i18n("compresses.minLength"),
    dataIndex: "minLength",
    sorter: (a, b) => a.minLength - b.minLength
  },
  {
    title: i18n("compresses.filter"),
    dataIndex: "filter"
  }
];

const fields = [
  {
    label: i18n("compresses.name"),
    key: "name",
    placeholder: i18n("compresses.namePlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("compresses.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("compresses.level"),
    key: "level",
    type: "number",
    placeholder: i18n("compresses.levelPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("compresses.levelRequireMessage")
      }
    ]
  },
  {
    label: i18n("compresses.minLength"),
    key: "minLength",
    type: "number",
    placeholder: i18n("compresses.minLengthPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("compresses.minLengthRequireMessage")
      }
    ]
  },
  {
    label: i18n("compresses.filter"),
    key: "filter",
    placeholder: i18n("compresses.filterPlaceHolder"),
    defaultValue: "text|javascript|json",
    rules: [
      {
        required: true,
        message: i18n("compresses.filterRequireMessage")
      }
    ]
  }
];

class Compresses extends Configs {
  constructor(props) {
    super(props);
    _.extend(this.state, {
      title: i18n("compresses.createUpdateTitle"),
      description: i18n("compresses.createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Compresses;
