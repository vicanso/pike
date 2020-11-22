///
/// 配置信息
/// 需要注意，由于后端返回配置有部分配置可能为空，如[]string hosts，
/// 自动生成的代码List.from时，对于为空的数据会导致异常，因此需要填充为List()
///
// ignore_for_file: argument_type_not_assignable
// ignore_for_file:  prefer_expression_function_bodies
import 'dart:convert';

import 'package:flutter/foundation.dart';

// fillEmptyList 填充空列表，用于在Config.fromMap中使用
void fillEmptyList(Map<String, dynamic> m) {
  m['locations']?.forEach((element) {
    [
      'prefixes',
      'rewrites',
      'respHeaders',
      'reqHeaders',
      'hosts',
    ].forEach((key) {
      element[key] ??= [];
    });
  });
  m['upstreams']?.forEach((element) {
    element['servers'] ??= [];
  });

  m['servers']?.forEach((element) {
    element['locations'] ??= [];
  });
}

class AdminConfig {
  final String user;
  final String password;
  final String prefix;
  AdminConfig({
    this.user,
    this.password,
    this.prefix,
  });

  AdminConfig copyWith({
    String user,
    String password,
    String prefix,
  }) {
    return AdminConfig(
      user: user ?? this.user,
      password: password ?? this.password,
      prefix: prefix ?? this.prefix,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'user': user,
      'password': password,
      'prefix': prefix,
    };
  }

  factory AdminConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return AdminConfig(
      user: map['user'],
      password: map['password'],
      prefix: map['prefix'],
    );
  }

  String toJson() => json.encode(toMap());

  factory AdminConfig.fromJson(String source) =>
      AdminConfig.fromMap(json.decode(source));

  @override
  String toString() =>
      'AdminConfig(user: $user, password: $password, prefix: $prefix)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is AdminConfig &&
        o.user == user &&
        o.password == password &&
        o.prefix == prefix;
  }

  @override
  int get hashCode => user.hashCode ^ password.hashCode ^ prefix.hashCode;
}

class CompressConfig {
  final String name;
  final Map<String, int> levels;
  final String remark;
  CompressConfig({
    this.name,
    this.levels,
    this.remark,
  });

  CompressConfig copyWith({
    String name,
    Map<String, int> levels,
    String remark,
  }) {
    return CompressConfig(
      name: name ?? this.name,
      levels: levels ?? this.levels,
      remark: remark ?? this.remark,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'levels': levels,
      'remark': remark,
    };
  }

  factory CompressConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return CompressConfig(
      name: map['name'],
      levels: Map<String, int>.from(map['levels']),
      remark: map['remark'],
    );
  }

  String toJson() => json.encode(toMap());

  factory CompressConfig.fromJson(String source) =>
      CompressConfig.fromMap(json.decode(source));

  @override
  String toString() =>
      'CompressConfig(name: $name, levels: $levels, remark: $remark)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is CompressConfig &&
        o.name == name &&
        mapEquals(o.levels, levels) &&
        o.remark == remark;
  }

  @override
  int get hashCode => name.hashCode ^ levels.hashCode ^ remark.hashCode;
}

class CacheConfig {
  final String name;
  final int size;
  final String hitForPass;
  final String remark;

  CacheConfig({
    this.name,
    this.size,
    this.hitForPass,
    this.remark,
  });

  CacheConfig copyWith({
    String name,
    int size,
    String hitForPass,
    String remark,
  }) {
    return CacheConfig(
      name: name ?? this.name,
      size: size ?? this.size,
      hitForPass: hitForPass ?? this.hitForPass,
      remark: remark ?? this.remark,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'size': size,
      'hitForPass': hitForPass,
      'remark': remark,
    };
  }

  factory CacheConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return CacheConfig(
      name: map['name'],
      size: map['size'],
      hitForPass: map['hitForPass'],
      remark: map['remark'],
    );
  }

  String toJson() => json.encode(toMap());

  factory CacheConfig.fromJson(String source) =>
      CacheConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'CacheConfig(name: $name, size: $size, hitForPass: $hitForPass, remark: $remark)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is CacheConfig &&
        o.name == name &&
        o.size == size &&
        o.hitForPass == hitForPass &&
        o.remark == remark;
  }

  @override
  int get hashCode {
    return name.hashCode ^
        size.hashCode ^
        hitForPass.hashCode ^
        remark.hashCode;
  }
}

class UpstreamServerConfig {
  final String addr;
  final bool backup;
  final bool healthy;
  UpstreamServerConfig({
    this.addr,
    this.backup,
    this.healthy,
  });

  UpstreamServerConfig copyWith({
    String addr,
    bool backup,
    bool healthy,
  }) {
    return UpstreamServerConfig(
      addr: addr ?? this.addr,
      backup: backup ?? this.backup,
      healthy: healthy ?? this.healthy,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'addr': addr,
      'backup': backup,
      'healthy': healthy,
    };
  }

  factory UpstreamServerConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return UpstreamServerConfig(
      addr: map['addr'],
      backup: map['backup'],
      healthy: map['healthy'],
    );
  }

  String toJson() => json.encode(toMap());

  factory UpstreamServerConfig.fromJson(String source) =>
      UpstreamServerConfig.fromMap(json.decode(source));

  @override
  String toString() =>
      'UpstreamServerConfig(addr: $addr, backup: $backup, healthy: $healthy)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is UpstreamServerConfig &&
        o.addr == addr &&
        o.backup == backup &&
        o.healthy == healthy;
  }

  @override
  int get hashCode => addr.hashCode ^ backup.hashCode ^ healthy.hashCode;
}

class UpstreamConfig {
  final String name;
  final String healthCheck;
  final String policy;
  final bool enableH2C;
  final String acceptEncoding;
  final List<UpstreamServerConfig> servers;
  final String remark;

  UpstreamConfig({
    this.name,
    this.healthCheck,
    this.policy,
    this.enableH2C,
    this.acceptEncoding,
    this.servers,
    this.remark,
  });

  UpstreamConfig copyWith({
    String name,
    String healthCheck,
    String policy,
    bool enableH2C,
    String acceptEncoding,
    List<UpstreamServerConfig> servers,
    String remark,
  }) {
    return UpstreamConfig(
      name: name ?? this.name,
      healthCheck: healthCheck ?? this.healthCheck,
      policy: policy ?? this.policy,
      enableH2C: enableH2C ?? this.enableH2C,
      acceptEncoding: acceptEncoding ?? this.acceptEncoding,
      servers: servers ?? this.servers,
      remark: remark ?? this.remark,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'healthCheck': healthCheck,
      'policy': policy,
      'enableH2C': enableH2C,
      'acceptEncoding': acceptEncoding,
      'servers': servers?.map((x) => x?.toMap())?.toList(),
      'remark': remark,
    };
  }

  factory UpstreamConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return UpstreamConfig(
      name: map['name'],
      healthCheck: map['healthCheck'],
      policy: map['policy'],
      enableH2C: map['enableH2C'],
      acceptEncoding: map['acceptEncoding'],
      servers: List<UpstreamServerConfig>.from(
          map['servers']?.map((x) => UpstreamServerConfig.fromMap(x))),
      remark: map['remark'],
    );
  }

  String toJson() => json.encode(toMap());

  factory UpstreamConfig.fromJson(String source) =>
      UpstreamConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'UpstreamConfig(name: $name, healthCheck: $healthCheck, policy: $policy, enableH2C: $enableH2C, acceptEncoding: $acceptEncoding, servers: $servers, remark: $remark)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is UpstreamConfig &&
        o.name == name &&
        o.healthCheck == healthCheck &&
        o.policy == policy &&
        o.enableH2C == enableH2C &&
        o.acceptEncoding == acceptEncoding &&
        listEquals(o.servers, servers) &&
        o.remark == remark;
  }

  @override
  int get hashCode {
    return name.hashCode ^
        healthCheck.hashCode ^
        policy.hashCode ^
        enableH2C.hashCode ^
        acceptEncoding.hashCode ^
        servers.hashCode ^
        remark.hashCode;
  }
}

class LocationConfig {
  final String name;
  final String upstream;
  final List<String> prefixes;
  final List<String> hosts;
  final List<String> rewrites;
  final List<String> respHeaders;
  final List<String> reqHeaders;
  final String proxyTimeout;
  final String remark;

  LocationConfig({
    this.name,
    this.upstream,
    this.prefixes,
    this.rewrites,
    this.respHeaders,
    this.reqHeaders,
    this.hosts,
    this.proxyTimeout,
    this.remark,
  });

  LocationConfig copyWith({
    String name,
    String upstream,
    List<String> prefixes,
    List<String> rewrites,
    List<String> respHeaders,
    List<String> reqHeaders,
    List<String> hosts,
    String proxyTimeout,
    String remark,
  }) {
    return LocationConfig(
      name: name ?? this.name,
      upstream: upstream ?? this.upstream,
      prefixes: prefixes ?? this.prefixes,
      rewrites: rewrites ?? this.rewrites,
      respHeaders: respHeaders ?? this.respHeaders,
      reqHeaders: reqHeaders ?? this.reqHeaders,
      hosts: hosts ?? this.hosts,
      proxyTimeout: proxyTimeout ?? this.proxyTimeout,
      remark: remark ?? this.remark,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'upstream': upstream,
      'prefixes': prefixes,
      'rewrites': rewrites,
      'respHeaders': respHeaders,
      'reqHeaders': reqHeaders,
      'hosts': hosts,
      'proxyTimeout': proxyTimeout,
      'remark': remark,
    };
  }

  factory LocationConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return LocationConfig(
      name: map['name'],
      upstream: map['upstream'],
      prefixes: List<String>.from(map['prefixes']),
      rewrites: List<String>.from(map['rewrites']),
      respHeaders: List<String>.from(map['respHeaders']),
      reqHeaders: List<String>.from(map['reqHeaders']),
      hosts: List<String>.from(map['hosts']),
      proxyTimeout: map['proxyTimeout'],
      remark: map['remark'],
    );
  }

  String toJson() => json.encode(toMap());

  factory LocationConfig.fromJson(String source) =>
      LocationConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'LocationConfig(name: $name, upstream: $upstream, prefixes: $prefixes, rewrites: $rewrites, respHeaders: $respHeaders, reqHeaders: $reqHeaders, hosts: $hosts, proxyTimeout: $proxyTimeout, remark: $remark)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is LocationConfig &&
        o.name == name &&
        o.upstream == upstream &&
        listEquals(o.prefixes, prefixes) &&
        listEquals(o.rewrites, rewrites) &&
        listEquals(o.respHeaders, respHeaders) &&
        listEquals(o.reqHeaders, reqHeaders) &&
        listEquals(o.hosts, hosts) &&
        o.proxyTimeout == proxyTimeout &&
        o.remark == remark;
  }

  @override
  int get hashCode {
    return name.hashCode ^
        upstream.hashCode ^
        prefixes.hashCode ^
        rewrites.hashCode ^
        respHeaders.hashCode ^
        reqHeaders.hashCode ^
        hosts.hashCode ^
        proxyTimeout.hashCode ^
        remark.hashCode;
  }
}

class ServerConfig {
  final String logFormat;
  final String addr;
  final List<String> locations;
  final String cache;
  final String compress;
  final String compressMinLength;
  final String compressContentTypeFilter;
  final String remark;

  ServerConfig({
    this.logFormat,
    this.addr,
    this.locations,
    this.cache,
    this.compress,
    this.compressMinLength,
    this.compressContentTypeFilter,
    this.remark,
  });

  ServerConfig copyWith({
    String logFormat,
    String addr,
    List<String> locations,
    String cache,
    String compress,
    String compressMinLength,
    String compressContentTypeFilter,
    String remark,
  }) {
    return ServerConfig(
      logFormat: logFormat ?? this.logFormat,
      addr: addr ?? this.addr,
      locations: locations ?? this.locations,
      cache: cache ?? this.cache,
      compress: compress ?? this.compress,
      compressMinLength: compressMinLength ?? this.compressMinLength,
      compressContentTypeFilter:
          compressContentTypeFilter ?? this.compressContentTypeFilter,
      remark: remark ?? this.remark,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'logFormat': logFormat,
      'addr': addr,
      'locations': locations,
      'cache': cache,
      'compress': compress,
      'compressMinLength': compressMinLength,
      'compressContentTypeFilter': compressContentTypeFilter,
      'remark': remark,
    };
  }

  factory ServerConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return ServerConfig(
      logFormat: map['logFormat'],
      addr: map['addr'],
      locations: List<String>.from(map['locations']),
      cache: map['cache'],
      compress: map['compress'],
      compressMinLength: map['compressMinLength'],
      compressContentTypeFilter: map['compressContentTypeFilter'],
      remark: map['remark'],
    );
  }

  String toJson() => json.encode(toMap());

  factory ServerConfig.fromJson(String source) =>
      ServerConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'ServerConfig(logFormat: $logFormat, addr: $addr, locations: $locations, cache: $cache, compress: $compress, compressMinLength: $compressMinLength, compressContentTypeFilter: $compressContentTypeFilter, remark: $remark)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is ServerConfig &&
        o.logFormat == logFormat &&
        o.addr == addr &&
        listEquals(o.locations, locations) &&
        o.cache == cache &&
        o.compress == compress &&
        o.compressMinLength == compressMinLength &&
        o.compressContentTypeFilter == compressContentTypeFilter &&
        o.remark == remark;
  }

  @override
  int get hashCode {
    return logFormat.hashCode ^
        addr.hashCode ^
        locations.hashCode ^
        cache.hashCode ^
        compress.hashCode ^
        compressMinLength.hashCode ^
        compressContentTypeFilter.hashCode ^
        remark.hashCode;
  }
}

class Config {
  final String yaml;
  final AdminConfig admin;
  final List<CompressConfig> compresses;
  final List<CacheConfig> caches;
  final List<UpstreamConfig> upstreams;
  final List<LocationConfig> locations;
  final List<ServerConfig> servers;
  Config({
    this.yaml,
    this.admin,
    this.compresses,
    this.caches,
    this.upstreams,
    this.locations,
    this.servers,
  });

  // validateForDelete 校验该字段是否可删除
  bool validateForDelete(String category, String name) {
    var valid = true;
    switch (category) {
      case 'compress':
        servers?.forEach((element) {
          if (element.compress == name) {
            valid = false;
          }
        });
        break;
      case 'cache':
        servers?.forEach((element) {
          if (element.cache == name) {
            valid = false;
          }
        });
        break;
      case 'upstream':
        locations?.forEach((element) {
          if (element.upstream == name) {
            valid = false;
          }
        });
        break;
      case 'location':
        servers?.forEach((element) {
          element.locations?.forEach((location) {
            if (location == name) {
              valid = false;
            }
          });
        });
        break;
      default:
        valid = false;
    }
    return valid;
  }

  Config copyWith({
    String yaml,
    AdminConfig admin,
    List<CompressConfig> compresses,
    List<CacheConfig> caches,
    List<UpstreamConfig> upstreams,
    List<LocationConfig> locations,
    List<ServerConfig> servers,
  }) {
    return Config(
      yaml: yaml ?? this.yaml,
      admin: admin ?? this.admin,
      compresses: compresses ?? this.compresses,
      caches: caches ?? this.caches,
      upstreams: upstreams ?? this.upstreams,
      locations: locations ?? this.locations,
      servers: servers ?? this.servers,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'yaml': yaml,
      'admin': admin?.toMap(),
      'compresses': compresses?.map((x) => x?.toMap())?.toList(),
      'caches': caches?.map((x) => x?.toMap())?.toList(),
      'upstreams': upstreams?.map((x) => x?.toMap())?.toList(),
      'locations': locations?.map((x) => x?.toMap())?.toList(),
      'servers': servers?.map((x) => x?.toMap())?.toList(),
    };
  }

  factory Config.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;
    fillEmptyList(map);
    return Config(
      yaml: map['yaml'],
      admin: AdminConfig.fromMap(map['admin']),
      compresses: List<CompressConfig>.from(
          map['compresses']?.map((x) => CompressConfig.fromMap(x))),
      caches: List<CacheConfig>.from(
          map['caches']?.map((x) => CacheConfig.fromMap(x))),
      upstreams: List<UpstreamConfig>.from(
          map['upstreams']?.map((x) => UpstreamConfig.fromMap(x))),
      locations: List<LocationConfig>.from(
          map['locations']?.map((x) => LocationConfig.fromMap(x))),
      servers: List<ServerConfig>.from(
          map['servers']?.map((x) => ServerConfig.fromMap(x))),
    );
  }

  String toJson() => json.encode(toMap());

  factory Config.fromJson(String source) => Config.fromMap(json.decode(source));

  @override
  String toString() {
    return 'Config(yaml: $yaml, admin: $admin, compresses: $compresses, caches: $caches, upstreams: $upstreams, locations: $locations, servers: $servers)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is Config &&
        o.yaml == yaml &&
        o.admin == admin &&
        listEquals(o.compresses, compresses) &&
        listEquals(o.caches, caches) &&
        listEquals(o.upstreams, upstreams) &&
        listEquals(o.locations, locations) &&
        listEquals(o.servers, servers);
  }

  @override
  int get hashCode {
    return yaml.hashCode ^
        admin.hashCode ^
        compresses.hashCode ^
        caches.hashCode ^
        upstreams.hashCode ^
        locations.hashCode ^
        servers.hashCode;
  }
}
