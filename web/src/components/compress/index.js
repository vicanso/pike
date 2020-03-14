import i18n from "../../i18n";
import Configs from "../configs";

const category = "compresses";

const columns = [
  {
    title: i18n("compress.name"),
    dataIndex: "name"
  },
  {
    title: i18n("compress.level"),
    dataIndex: "level",
    sorter: (a, b) => a.level - b.level
  },
  {
    title: i18n("compress.minLength"),
    dataIndex: "minLength",
    sorter: (a, b) => a.minLength - b.minLength
  },
  {
    title: i18n("compress.filter"),
    dataIndex: "filter"
  },
  {
    title: i18n("common.description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: i18n("compress.name"),
    key: "name",
    placeholder: i18n("compress.namePlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("compress.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("compress.level"),
    key: "level",
    type: "number",
    placeholder: i18n("compress.levelPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("compress.levelRequireMessage")
      }
    ]
  },
  {
    label: i18n("compress.minLength"),
    key: "minLength",
    type: "number",
    placeholder: i18n("compress.minLengthPlaceHolder"),
    rules: [
      {
        required: true,
        message: i18n("compress.minLengthRequireMessage")
      }
    ]
  },
  {
    label: i18n("compress.filter"),
    key: "filter",
    placeholder: i18n("compress.filterPlaceHolder"),
    defaultValue: "text|javascript|json",
    rules: [
      {
        required: true,
        message: i18n("compress.filterRequireMessage")
      }
    ]
  },
  {
    label: i18n("common.description"),
    key: "description",
    type: "textarea",
    placeholder: i18n("common.descriptionPlaceholder")
  }
];

class Compresses extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: i18n("compress.createUpdateTitle"),
      description: i18n("compress.createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Compresses;
