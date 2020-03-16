import Configs from "../configs";
import { getCacheI18n, getCommonI18n } from "../../i18n";

const category = "caches";

const columns = [
  {
    title: getCacheI18n("name"),
    dataIndex: "name"
  },
  {
    title: getCacheI18n("size"),
    dataIndex: "size",
    sorter: (a, b) => a.size - b.size
  },
  {
    title: getCacheI18n("zone"),
    dataIndex: "zone",
    sorter: (a, b) => a.zone - b.zone
  },
  {
    title: getCacheI18n("hitForPass"),
    dataIndex: "hitForPass",
    sorter: (a, b) => a.hitForPass - b.hitForPass
  },
  {
    title: getCommonI18n("description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: getCacheI18n("name"),
    key: "name",
    placeholder: getCacheI18n("namePlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("size"),
    key: "size",
    type: "number",
    placeholder: getCacheI18n("sizePlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("sizeRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("zone"),
    key: "zone",
    type: "number",
    placeholder: getCacheI18n("zonePlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("zoneRequireMessage")
      }
    ]
  },
  {
    label: getCacheI18n("hitForPass"),
    key: "hitForPass",
    type: "number",
    placeholder: getCacheI18n("hitForPassPlaceholder"),
    rules: [
      {
        required: true,
        message: getCacheI18n("hitForPassRequireMessage")
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

class Caches extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: getCacheI18n("createUpdateTitle"),
      description: getCacheI18n("createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Caches;
