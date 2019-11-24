import _ from "lodash";

import i18n from "../../i18n";
import Configs from "../configs";

const category = "caches";

const columns = [
  {
    title: i18n("caches.name"),
    dataIndex: "name"
  },
  {
    title: i18n("caches.size"),
    dataIndex: "size",
    sorter: (a, b) => a.size - b.size
  },
  {
    title: i18n("caches.zone"),
    dataIndex: "zone",
    sorter: (a, b) => a.zone - b.zone
  },
  {
    title: i18n("caches.hitForPass"),
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
    label: i18n("caches.name"),
    key: "name",
    placeholder: i18n("caches.namePlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("caches.nameRequireMessage")
      }
    ]
  },
  {
    label: i18n("caches.zone"),
    key: "zone",
    type: "number",
    placeholder: i18n("caches.zonePlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("caches.zoneRequireMessage")
      }
    ]
  },
  {
    label: i18n("caches.size"),
    key: "size",
    type: "number",
    placeholder: i18n("caches.sizePlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("caches.sizeRequireMessage")
      }
    ]
  },
  {
    label: i18n("caches.hitForPass"),
    key: "hitForPass",
    type: "number",
    placeholder: i18n("caches.hitForPassPlaceholder"),
    rules: [
      {
        required: true,
        message: i18n("caches.hitForPassRequireMessage")
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
    _.extend(this.state, {
      title: i18n("caches.createUpdateTitle"),
      description: i18n("caches.createUpdateDescription"),
      category,
      columns,
      fields
    });
  }
}

export default Caches;
