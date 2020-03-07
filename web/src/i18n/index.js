import _ from "lodash";

const key = "language";

const lang = localStorage.getItem(key) || "en";

const navEn = {
  caches: "Caches",
  compresses: "Compresses",
  upstreams: "Upstreams",
  locations: "Locations",
  servers: "Servers",
  admin: "Admin"
};
const navZh = {
  caches: "缓存",
  compresses: "压缩",
  upstreams: "Upstreams",
  locations: "Locations",
  servers: "HTTP服务器",
  admin: "管理配置"
};

const commonEn = {
  lang: "Language",
  second: "s",
  minute: "m",
  hour: "h",
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
  lang: "语言",
  second: "秒",
  minute: "分",
  hour: "时",
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
    "Create or update http cache for pike, the max size of cache is zone * size, 256 * 1000 is suitable for most website. Hit for pass is the ttl for cache's pass status, 300 seconds(5 minutes) is suitable.",
  name: "Name",
  namePlaceholder: "Please input the cache's name, only support alphabets",
  nameRequireMessage: "The cache's name can't be empty!",
  zone: "Zone Size",
  zonePlaceholder: "Please input the cache's zone size",
  zoneRequireMessage: "The cache's zone size should be gt 0",
  size: "Size",
  sizePlaceholder: "Please input the cache's size, 256 is suitable.",
  sizeRequireMessage: "The cache's size should be gt 0",
  hitForPass: "Hit For Pass",
  hitForPassPlaceholder: "Please input hit for pass ttl for cache",
  hitForPassRequireMessage: "The cache's hit for pass should be gt 0"
};
const cacheZh = {
  createUpdateTitle: "创建或更新缓存",
  createUpdateDescription:
    "创建或更新HTTP缓存，缓存的最大长度是 zone * size，256 * 1000已适用于大部分网站。Hit for pass是缓存pass状态的有效期，300秒（5分钟）是比较适合的值。",
  name: "名称",
  namePlaceholder: "请输入缓存的名称，只支持字母",
  nameRequireMessage: "缓存的名称不能为空",
  zone: "空间大小",
  zonePlaceholder: "请输入缓存空间的长度",
  zoneRequireMessage: "缓存空间的长度必须大于0",
  size: "大小",
  sizePlaceholder: "请输入缓存的长度，建议设置为256",
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
  namePlaceHolder: "Please input the compress's name, only support alphabets",
  nameRequireMessage: "The compress's name can't be empty!",
  level: "Level",
  levelPlaceHolder: "Please input the compress's level, 9 is better",
  levelRequireMessage: "The compress level can't be empty!",
  minLength: "Min Length",
  minLengthPlaceHolder:
    "Please input the min byte's length to compress, 1024 is better",
  minLengthRequireMessage: "The min length can't be empty!",
  filter: "Filter",
  filterPlaceHolder:
    "Please input the regexp for check content type to compress",
  filterRequireMessage: "The content type filter can't be empty!"
};
const compressZh = {
  createUpdateTitle: "创建或更新配置缓存",
  createUpdateDescription:
    "指定HTTP压缩的级别，可限定最小压缩长度以及压缩数据类型。",
  name: "名称",
  namePlaceHolder: "请输入压缩配置的名称，仅支持字母",
  nameRequireMessage: "压缩配置的名称不能为空",
  level: "压缩等级",
  levelPlaceHolder: "请输入压缩的级别，9为比较合适的压缩级别",
  levelRequireMessage: "压缩级别不能为空",
  minLength: "压缩最小长度",
  minLengthPlaceHolder: "请输入的最小字节长度，1024为比较合适的最小长度",
  minLengthRequireMessage: "最小字节长度不能为空",
  filter: "筛选",
  filterPlaceHolder: "请输入对响应内容筛选的正式表达式，默认为text|json",
  filterRequireMessage: "内容类型筛选不能为空"
};

const upstreamEn = {
  createUpdateTitle: "Create or update upstream",
  createUpdateDescription:
    "Set the upstream's address list for location's proxy, the policy of choosing upstream server, and the health check path.",
  name: "Name",
  namePlaceHolder: "Please input the name of upstream, only support alphabets",
  nameRequireMessage: "The name of upstream can't be empty!",
  policy: "Policy",
  policyPlaceHolder: "Please select the policy of chosing upstream server",
  servers: "Servers",
  serversRequireMessage: "The servers of upstream can't be empty!",
  serverAddrPlaceHolder:
    "Please input the addreass of upstream server, eg: http://127.0.0.1:3000.",
  serverAddrRequireMessage: "The address of upstream server can't be empty!",
  backup: "Backup",
  healthCheck: "Health Check",
  healthCheckPlaceHolder: "Please input the url path of health check, eg: /ping"
};
const upstreamZh = {
  createUpdateTitle: "创建或更新Upstream",
  createUpdateDescription:
    "设置upstream服务的服务地址列表，相关的选择策略以及健康检测配置。",
  name: "名称",
  namePlaceHolder: "请输入upstream的名称，仅支持字母",
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
  namePlaceHolder: "Please input the name of location, only support alphabets",
  nameRequireMessage: "The name of location can't be empty!",
  upstream: "Upstream",
  upstreamPlaceHolder: "Please select the upstream of location",
  upstreamRequireMessage: "The upstream of location can't be empty!",
  hosts: "Hosts",
  hostsPlaceHolder: "Please input the host for location, optional",
  prefixs: "Prefixs",
  prefixsPlaceHolder: "Please input the prefix for location, optional",
  rewrites: "URL Rewrites",
  rewriteOriginalPlaceHolder: "Please input the original url, eg: /api/*",
  rewriteNewPlaceHolder: "Please input the rewrite url, eg: /$1",
  reqHeader: "Request Header",
  resHeader: "Response Header",
  headerNamePlaceHolder: "Please input the header's name, eg: X-Request-ID",
  headerValuePlaceHolder: "Please input the header's value eg: 1001"
};
const locationZh = {
  createUpdateTitle: "创建或更新location",
  createUpdateDescription:
    "创建或更新用于HTTP服务的location，包括host列表，url前缀列表，upstream、请求头与响应头等。",
  name: "名称",
  namePlaceHolder: "请输入location的名称，仅支持字母",
  nameRequireMessage: "location的名称不能为空！",
  upstream: "Upstream",
  upstreamPlaceHolder: "请选择该location的upstream",
  upstreamRequireMessage: "该location的upstream不能为空！",
  hosts: "Hosts",
  hostsPlaceHolder: "请输入该location使用的host，可选",
  prefixs: "前缀",
  prefixsPlaceHolder: "请输入该location的URL前缀，可选",
  rewrites: "URL重写",
  rewriteOriginalPlaceHolder: "请输入原始URL，如：/api/*",
  rewriteNewPlaceHolder: "请输入重写的URL，如：/$1",
  reqHeader: "请求头",
  resHeader: "响应头",
  headerNamePlaceHolder: "请输入HTTP头的名称，如：X-Request-ID",
  headerValuePlaceHolder: "请输入HTTP头的值，如：1001"
};

const serverEn = {
  createUpdateTitle: "Create or update http server",
  createUpdateDescription:
    "Create or update http server, the listen address and port shouldn't be used.",
  name: "Name",
  namePlaceHolder:
    "Please input the name of http server, only support alphabets",
  nameRequireMessage: "The name of http server can't be empty!",
  cache: "Cache",
  cachePlaceHolder: "Please select the cache config for http server",
  cacheRequireMessage: "The cache config for http server can't be empty!",
  compress: "Compress",
  compressPlaceHolder: "Please select the compress config for http server",
  compressRequireMessage: "The compress config for http server can't be empty!",
  locations: "Locations",
  locationsPlaceHolder: "Please select the locations config for http server",
  locationsRequireMesage:
    "The locations config for http server can't be empty!",
  etag: "ETag",
  addr: "Address",
  addrPlaceHolder:
    "Please input the listen address for http server, eg: :7000 or 127.0.0.1:7000",
  addrRequireMessage: "The listen address can't be empty!",
  concurrency: "Concurrency",
  concurrencyPlaceHolder: "Please input the limit concurrency",
  readTimeout: "ReadTimeout",
  readTimeoutPlaceHolder: "Please input the read timeout",
  writeTimeout: "WriteTimeout",
  writeTimeoutPlaceHolder: "Please input the write timeout",
  idleTimeout: "IdleTimeout",
  idleTimeoutPlaceHolder: "Please input the idle timeout",
  maxHeaderBytes: "MaxHeaderBytes",
  maxHeaderBytesPlaceHolder: "Please input the max header bytes limit"
};
const serverZh = {
  createUpdateTitle: "创建或更新HTTP服务器",
  createUpdateDescription:
    "创建或更新HTTP服务器，其中监听的地址与端口必须未被使用的。",
  name: "名称",
  namePlaceHolder: "请输入HTTP服务器的名称，仅支持字母",
  nameRequireMessage: "HTTP服务器的名称不能为空！",
  cache: "缓存",
  cachePlaceHolder: "请选择HTTP服务器使用的缓存配置",
  cacheRequireMessage: "HTTP服务器的缓存配置不能为空!",
  compress: "压缩",
  compressPlaceHolder: "请选择HTTP服务器使用的压缩配置",
  compressRequireMessage: "HTTP服务器的压缩配置不能为空！",
  locations: "locations",
  locationsPlaceHolder: "请选择HTTP服务器使用的locations配置",
  locationsRequireMesage: "HTTP服务器的locations配置不能为空",
  etag: "ETag",
  addr: "监听地址",
  addrPlaceHolder: "请输入HTTP服务器的监听地址，如：:7000 或者 127.0.0.1:7000",
  addrRequireMessage: "HTTP服务器监听地址不能为空！",
  concurrency: "并发数",
  concurrencyPlaceHolder: "请输入并发数限制",
  readTimeout: "读超时",
  readTimeoutPlaceHolder: "请输入读超时参数",
  writeTimeout: "写超时",
  writeTimeoutPlaceHolder: "请输入写超时参数",
  idleTimeout: "空闲超时",
  idleTimeoutPlaceHolder: "请输入空闲超时参数",
  maxHeaderBytes: "最大请求头长度",
  maxHeaderBytesPlaceHolder: "请输入最大请求头长度限制参数"
};

const adminEn = {
  createUpdateTitle: "Create or update admin config",
  createUpdateDescription:
    "Create or update admin config for http basic auth. FYA, only one config is effective.",
  user: "User",
  userPlaceHolder: "Please input the account",
  password: "Password",
  passwordPlaceHolder: "Please input the password for user",
  prefix: "URL Prefix",
  prefixPlaceHolder: "Please input the url prefix for admin",
  prefixRequireMessage: "URL prefix for admin can't be empty!",
  enabledInternetAccess: "Internet Access"
};
const adminZh = {
  createUpdateTitle: "创建或更新管理配置",
  createUpdateDescription:
    "创建或更新用于HTTP基础认证的相关配置，请注意管理配置只有一个生效。",
  user: "用户",
  userPlaceHolder: "请输入用户名",
  password: "密码",
  passwordPlaceHolder: "请输入用户密码",
  prefix: "地址前缀",
  prefixPlaceHolder: "请输入管理后台的地址前缀",
  prefixRequireMessage: "管理后台地址前缀不能为空！",
  enabledInternetAccess: "外网访问"
};

const i18ns = {
  en: {
    common: commonEn,
    nav: navEn,
    cache: cacheEn,
    compress: compressEn,
    upstream: upstreamEn,
    location: locationEn,
    server: serverEn,
    admin: adminEn
  },
  zh: {
    common: commonZh,
    nav: navZh,
    cache: cacheZh,
    compress: compressZh,
    upstream: upstreamZh,
    location: locationZh,
    server: serverZh,
    admin: adminZh
  }
};

export default field => {
  return _.get(i18ns, `${lang}.${field}`) || "";
};

export function changeToEnglish() {
  localStorage.setItem(key, "en");
}
export function changeToChinese() {
  localStorage.setItem(key, "zh");
}
