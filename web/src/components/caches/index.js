import i18n from "../../i18n";
import Configs from "../configs";

const category = "caches";

const columns = [
  {
    title: i18n("cache.name"),
    dataIndex: "name"
  },
  {
    title: i18n("cache.size"),
    dataIndex: "size",
    sorter: (a, b) => a.size - b.size
  },
  {
    title: i18n("cache.zone"),
    dataIndex: "zone",
    sorter: (a, b) => a.zone - b.zone
  },
  {
    title: i18n("cache.hitForPass"),
    dataIndex: "hitForPass",
    sorter: (a, b) => a.hitForPass - b.hitForPass
  },
  {
    title: i18n("common.description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: i18n("cache.name"),
    key: "name",
    placeholder: i18n("cache.namePlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("cache.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("cache.size"),
    key: "size",
    type: "number",
    placeholder: i18n("cache.sizePlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("cache.sizeRequireMessage")
      }
    ]
  },
  {
    label: i18n("cache.zone"),
    key: "zone",
    type: "number",
    placeholder: i18n("cache.zonePlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("cache.zoneRequireMessage")
      }
    ]
  },
  {
    label: i18n("cache.hitForPass"),
    key: "hitForPass",
    type: "number",
    placeholder: i18n("cache.hitForPassPlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("cache.hitForPassRequireMessage")
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

class Caches extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: i18n("cache.createUpdateTitle"),
      description: i18n("cache.createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Caches;
