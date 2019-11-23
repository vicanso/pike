import _ from "lodash";

let lang = "en";

const navEn = {
  caches: "Caches"
};
const navZh = {
  caches: "缓存"
};

const commonEn = {
  action: "Action",
  description: "Description",
  descriptionPlaceholder: "Please input the description",
  add: "Add",
  submit: "Submit",
  update: "Update",
  back: "Back",
  delete: "Delete",
  deleteTips: "Are you sure to delete this config?"
};
const commonZh = {
  action: "操作",
  description: "描述",
  descriptionPlaceholder: "请输入描述",
  add: "添加",
  submit: "提交",
  update: "更新",
  back: "返回",
  delete: "删除",
  deleteTips: "确定要删除此配置吗？"
};

// 缓存相关文本配置
const cachesEn = {
  createUpdateTitle: "Create or update cache",
  createUpdateDescription:
    "Create or update http cache for pike, the cache's size is zone * size, 1000 * 1000 is suitable for most website. Hit for pass is the ttle for cache's pass status, 300 seconds(5 minutes) is suitable.",
  name: "Name",
  namePlaceholder: "Please input the cache's name",
  nameRequireMessage: "The cache's name can't be empty!",
  zone: "Zone Size",
  zonePlaceholder: "Please input the cache's zone size",
  zoneRequireMessage: "The cache's zone size should be gt 0",
  size: "Size",
  sizePlaceholder: "Please input the cache's size",
  sizeRequireMessage: "The cache's size should be gt 0",
  hitForPass: "Hit For Pass",
  hitForPassPlaceholder: "Please input hit for pass ttl for cache",
  hitForPassRequireMessage: "The cache's hit for pass should be gt 0"
};
const cachesZh = {
  createTitle: "创建或更新缓存",
  createUpdateDescription:
    "创建或更新HTTP缓存，缓存的大小由 zone * size，1000 * 1000已适用于大部分网站。Hit for pass是缓存pass状态的有效期，300秒（5分钟）是比较适合的值。",
  name: "名称",
  namePlaceholder: "缓存的名称",
  nameRequireMessage: "请输入缓存的名称!",
  zone: "空间大小",
  zonePlaceholder: "请输入缓存空间的长度",
  zoneRequireMessage: "缓存空间的长度必须大于0",
  size: "大小",
  sizePlaceholder: "请输入缓存的长度",
  sizeRequireMessage: "缓存的长度必须大于0",
  hitForPass: "Hit For Pass",
  hitForPassPlaceholder: "请输入hit for pass的有效期",
  hitForPassRequireMessage: "hit for pass的有效期必须大于0"
};

const i18ns = {
  en: {
    common: commonEn,
    nav: navEn,
    caches: cachesEn
  },
  zh: {
    common: commonZh,
    nav: navZh,
    caches: cachesZh
  }
};

export default field => {
  return _.get(i18ns, `${lang}.${field}`) || "";
};
