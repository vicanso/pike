import Configs from "../configs";
import { getCertI18n, getCommonI18n } from "../../i18n";

const category = "certs";

const renderFile = row => {
  if (!row) {
    return;
  }
  return `data:${row.substring(0, 8)}... length:${row.length}`;
};

const columns = [
  {
    title: getCertI18n("name"),
    dataIndex: "name"
  },
  {
    title: getCertI18n("key"),
    dataIndex: "key",
    render: renderFile
  },
  {
    title: getCertI18n("cert"),
    dataIndex: "cert",
    render: renderFile
  },
  {
    title: getCommonI18n("description"),
    dataIndex: "description"
  }
];

const fields = [
  {
    label: getCertI18n("name"),
    key: "name",
    placeholder: getCertI18n("namePlaceHolder"),
    rules: [
      {
        required: true,
        message: getCertI18n("nameRequireMessage")
      }
    ]
  },
  {
    label: getCertI18n("key"),
    key: "key",
    type: "upload",
    placeholder: getCommonI18n("upload"),
    rules: [
      {
        required: true,
        message: getCertI18n("fileRequireMessage")
      }
    ]
  },
  {
    label: getCertI18n("cert"),
    key: "cert",
    type: "upload",
    placeholder: getCommonI18n("upload"),
    rules: [
      {
        required: true,
        message: getCertI18n("fileRequireMessage")
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

class Certs extends Configs {
  constructor(props) {
    super(props);
    Object.assign(this.state, {
      title: getCertI18n("title"),
      columns,
      fields,
      category
    });
  }
}

export default Certs;
