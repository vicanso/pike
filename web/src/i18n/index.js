import _ from "lodash";

let lang = "en";

const navEn = {
  caches: "Caches",
  compresses: "Compresses",
  upstreams: "Upstreams",
  locations: "Locations"
};
const navZh = {
  caches: "缓存",
  compresses: "压缩",
  upstreams: "Upstreams",
  locations: "Locations"
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
const cacheEn = {
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
const cacheZh = {
  createUpdateTitle: "创建或更新缓存",
  createUpdateDescription:
    "创建或更新HTTP缓存，缓存的大小由 zone * size，1000 * 1000已适用于大部分网站。Hit for pass是缓存pass状态的有效期，300秒（5分钟）是比较适合的值。",
  name: "名称",
  namePlaceholder: "缓存的名称",
  nameRequireMessage: "缓存的名称不能为空",
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

const compressEn = {
  createUpdateTitle: "Create or update compress",
  createUpdateDescription:
    "Set the compress level, min compress byte's length and compress data content type.",
  name: "Name",
  namePlaceHolder: "Please input the compress's name",
  nameRequireMessage: "The compress's name can't be empty!",
  level: "Level",
  levelPlaceHolder: "Please input the compress's level",
  levelRequireMessage: "The compress level can't be empty!",
  minLength: "Min Length",
  minLengthPlaceHolder: "Please input the min byte's length to compress",
  minLengthRequireMessage: "The min length can't be empty!",
  filter: "Filter",
  filterPlaceHolder:
    "Please input the regexp for check content type to compress",
  filterRequireMessage: "The content type filter can't be empty!"
};
const compressZh = {
  createUpdateTitle: "创建或更新配置缓存",
  createUpdateDescription:
    "指定HTTP压缩的级别，可限定最少压缩长度以及压缩数据类型。",
  name: "名称",
  namePlaceHolder: "请输入压缩配置的名称",
  nameRequireMessage: "压缩配置的名称不能为空",
  level: "压缩等级",
  levelPlaceHolder: "请输入压缩的级别",
  levelRequireMessage: "压缩级别不能为空",
  minLength: "压缩最少长度",
  minLengthPlaceHolder: "请输入的最少字节长度",
  minLengthRequireMessage: "最少字节长度不能为空",
  filter: "筛选",
  filterPlaceHolder: "请输入对响应内容筛选的正式表达式，默认为text|json",
  filterRequireMessage: "内容类型筛选不能为空"
};

const upstreamEn = {
  createUpdateTitle: "Create or update upstream",
  createUpdateDescription:
    "Set the upstream's address list for location's proxy, the policy of choosing upstream server, and the health check path.",
  name: "Name",
  namePlaceHolder: "Please input the name of upstream",
  nameRequireMessage: "The name of upstream can't be empty!",
  policy: "Policy",
  policyPlaceHolder: "Please select the policy of chosing upstream server",
  servers: "Servers",
  serversRequireMessage: "The servers of upstream can't be empty!",
  serverAddrPlaceHolder:
    "Please input the addreass of upstream server, e.g.: http://127.0.0.1:3000.",
  serverAddrRequireMessage: "The address of upstream server can't be empty!",
  backup: "Backup",
  healthCheck: "Health Check",
  healthCheckPlaceHolder:
    "Please input the url path of health check, e.g.: /ping"
};
const upstreamZh = {
  createUpdateTitle: "创建或更新Upstream",
  createUpdateDescription:
    "设置upstream服务的服务地址列表，相关的选择策略以及健康检测配置。",
  name: "名称",
  namePlaceHolder: "请输入upstream的名称",
  nameRequireMessage: "upstream的名称不能为空！",
  policy: "策略",
  policyPlaceHolder: "请选择upstream的选择策略",
  servers: "服务列表",
  serversRequireMessage: "服务器列表不能为空！",
  addr: "地址",
  serverAddrPlaceHolder: "请输入upstream服务的地址，如：http://127.0.0.1:3000",
  serverAddrRequireMessage: "upstream服务的地址不能为空！",
  backup: "备用",
  healthCheck: "健康检测",
  healthCheckPlaceHolder: "请输入健康检测的路径，如： /ping"
};

const locationEn = {
  createUpdateTitle: "Create or update location",
  createUpdateDescription:
    "Create or update location for http server, include hosts, prefixs, upstream, request header and response header.",
  name: "Name",
  namePlaceHolder: "Please input the name of location",
  nameRequireMessage: "The name of location can't be empty!",
  upstream: "Upstream",
  upstreamPlaceHolder: "Please select the upstream of location",
  upstreamRequireMessage: "The upstream of location can't be empty!",
  hosts: "Hosts",
  hostsPlaceHolder: "Please input the host for location, optional",
  prefixs: "Prefixs",
  prefixsPlaceHolder: "Please input the prefix for location, optional",
  rewrites: "URL Rewrites",
  rewriteOriginalPlaceHolder: "Please input the original url",
  rewriteNewPlaceHolder: "Please input the rewrite url",
  reqHeader: "Request Header",
  resHeader: "Response Header",
  headerNamePlaceHolder: "Please input the header's name",
  headerValuePlaceHolder: "Please input the header's value"
};
const locationZh = {
  createUpdateTitle: "创建或更新location",
  createUpdateDescription:
    "创建或更新用于HTTP服务的location，包括host列表，url前缀列表，upstream、请求头与响应头等。",
  name: "名称",
  namePlaceHolder: "请输入location的名称",
  nameRequireMessage: "location的名称不能为空！",
  upstream: "Upstream",
  upstreamPlaceHolder: "请选择该location的upstream",
  upstreamRequireMessage: "该location的upstream不能为空！",
  hosts: "Hosts",
  hostsPlaceHolder: "请输入该location使用的host，可选",
  prefixs: "前缀",
  prefixsPlaceHolder: "请输入该location的URL前缀，可选",
  rewrites: "URL重写",
  rewriteOriginalPlaceHolder: "请输入原始URL",
  rewriteNewPlaceHolder: "请输入重写的URL",
  reqHeader: "请求头",
  resHeader: "响应头",
  headerNamePlaceHolder: "请输入HTTP头的名称",
  headerValuePlaceHolder: "请输入HTTP头的值"
};

const i18ns = {
  en: {
    common: commonEn,
    nav: navEn,
    cache: cacheEn,
    compress: compressEn,
    upstream: upstreamEn,
    location: locationEn
  },
  zh: {
    common: commonZh,
    nav: navZh,
    cache: cacheZh,
    compress: compressZh,
    upstream: upstreamZh,
    location: locationZh
  }
};

export default field => {
  return _.get(i18ns, `${lang}.${field}`) || "";
};
