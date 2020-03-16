import Configs from "../configs";
import { getCompressI18n, getCommonI18n } from "../../i18n";

const category = "compresses";

const columns = [
  {
    title: getCompressI18n("name"),
    dataIndex: "name"
  },
  {
    title: getCompressI18n("level"),
    dataIndex: "level",
    sorter: (a, b) => a.level - b.level
  },
  {
    title: getCompressI18n("minLength"),
    dataIndex: "minLength",
    sorter: (a, b) => a.minLength - b.minLength
  },
  {
    title: getCompressI18n("filter"),
    dataIndex: "filter"
  },
  {
    title: getCommonI18n("description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: getCompressI18n("name"),
    key: "name",
    placeholder: getCompressI18n("namePlaceHolder"),
    rules: [
      {
        required: true,
        message: getCompressI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getCompressI18n("level"),
    key: "level",
    type: "number",
    placeholder: getCompressI18n("levelPlaceHolder"),
    rules: [
      {
        required: true,
        message: getCompressI18n("levelRequireMessage")
      }
    ]
  },
  {
    label: getCompressI18n("minLength"),
    key: "minLength",
    type: "number",
    placeholder: getCompressI18n("minLengthPlaceHolder"),
    rules: [
      {
        required: true,
        message: getCompressI18n("minLengthRequireMessage")
      }
    ]
  },
  {
    label: getCompressI18n("filter"),
    key: "filter",
    placeholder: getCompressI18n("filterPlaceHolder"),
    defaultValue: "text|javascript|json",
    rules: [
      {
        required: true,
        message: getCompressI18n("filterRequireMessage")
      }
    ]
  },
  {
    label: getCommonI18n("description"),
    key: "description",
    type: "textarea",
    placeholder: getCommonI18n("descriptionPlaceholder")
  }
];

class Compresses extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: getCompressI18n("createUpdateTitle"),
      description: getCompressI18n("createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Compresses;
