///
/// 配置信息
/// 需要注意，由于后端返回配置有部分配置可能为空，如[]string hosts，
/// 自动生成的代码List.from时，对于为空的数据会导致异常，因此需要填充为List()
///
import 'dart:convert';

import 'package:flutter/foundation.dart';

// fillEmptyList 填充空列表，用于在Config.fromMap中使用
void fillEmptyList(Map<String, dynamic> m) {
  m['locations']?.forEach((e) {
    [
      'prefixes',
      'rewrites',
      'respHeaders',
      'reqHeaders',
      'hosts',
    ].forEach((key) {
      if (e[key] == null) {
        e[key] = List();
      }
    });
  });

  m['servers']?.forEach((e) {
    if (e['locations'] == null) {
      e['locations'] = List();
    }
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
  final Map levels;
  CompressConfig({
    this.name,
    this.levels,
  });

  CompressConfig copyWith({
    String name,
    Map levels,
  }) {
    return CompressConfig(
      name: name ?? this.name,
      levels: levels ?? this.levels,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'levels': levels,
    };
  }

  factory CompressConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return CompressConfig(
      name: map['name'],
      levels: Map.from(map['levels']),
    );
  }

  String toJson() => json.encode(toMap());

  factory CompressConfig.fromJson(String source) =>
      CompressConfig.fromMap(json.decode(source));

  @override
  String toString() => 'CompressConfig(name: $name, levels: $levels)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is CompressConfig && o.name == name && mapEquals(o.levels, levels);
  }

  @override
  int get hashCode => name.hashCode ^ levels.hashCode;
}

class CacheConfig {
  final String name;
  final int size;
  final String hitForPass;
  CacheConfig({
    this.name,
    this.size,
    this.hitForPass,
  });

  CacheConfig copyWith({
    String name,
    int size,
    String hitForPass,
  }) {
    return CacheConfig(
      name: name ?? this.name,
      size: size ?? this.size,
      hitForPass: hitForPass ?? this.hitForPass,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'size': size,
      'hitForPass': hitForPass,
    };
  }

  factory CacheConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return CacheConfig(
      name: map['name'],
      size: map['size'],
      hitForPass: map['hitForPass'],
    );
  }

  String toJson() => json.encode(toMap());

  factory CacheConfig.fromJson(String source) =>
      CacheConfig.fromMap(json.decode(source));

  @override
  String toString() =>
      'CacheConfig(name: $name, size: $size, hitForPass: $hitForPass)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is CacheConfig &&
        o.name == name &&
        o.size == size &&
        o.hitForPass == hitForPass;
  }

  @override
  int get hashCode => name.hashCode ^ size.hashCode ^ hitForPass.hashCode;
}

class UpstreamServerConfig {
  final String addr;
  final bool backup;
  UpstreamServerConfig({
    this.addr,
    this.backup,
  });

  UpstreamServerConfig copyWith({
    String addr,
    bool backup,
  }) {
    return UpstreamServerConfig(
      addr: addr ?? this.addr,
      backup: backup ?? this.backup,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'addr': addr,
      'backup': backup,
    };
  }

  factory UpstreamServerConfig.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return UpstreamServerConfig(
      addr: map['addr'],
      backup: map['backup'],
    );
  }

  String toJson() => json.encode(toMap());

  factory UpstreamServerConfig.fromJson(String source) =>
      UpstreamServerConfig.fromMap(json.decode(source));

  @override
  String toString() => 'UpstreamServerConfig(addr: $addr, backup: $backup)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is UpstreamServerConfig && o.addr == addr && o.backup == backup;
  }

  @override
  int get hashCode => addr.hashCode ^ backup.hashCode;
}

class UpstreamConfig {
  final String name;
  final String healthCheck;
  final String policy;
  final bool enableH2C;
  final String acceptEncoding;
  final List<UpstreamServerConfig> servers;
  UpstreamConfig({
    this.name,
    this.healthCheck,
    this.policy,
    this.enableH2C,
    this.acceptEncoding,
    this.servers,
  });

  UpstreamConfig copyWith({
    String name,
    String healthCheck,
    String policy,
    bool enableH2C,
    String acceptEncoding,
    List<UpstreamServerConfig> servers,
  }) {
    return UpstreamConfig(
      name: name ?? this.name,
      healthCheck: healthCheck ?? this.healthCheck,
      policy: policy ?? this.policy,
      enableH2C: enableH2C ?? this.enableH2C,
      acceptEncoding: acceptEncoding ?? this.acceptEncoding,
      servers: servers ?? this.servers,
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
    );
  }

  String toJson() => json.encode(toMap());

  factory UpstreamConfig.fromJson(String source) =>
      UpstreamConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'UpstreamConfig(name: $name, healthCheck: $healthCheck, policy: $policy, enableH2C: $enableH2C, acceptEncoding: $acceptEncoding, servers: $servers)';
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
        listEquals(o.servers, servers);
  }

  @override
  int get hashCode {
    return name.hashCode ^
        healthCheck.hashCode ^
        policy.hashCode ^
        enableH2C.hashCode ^
        acceptEncoding.hashCode ^
        servers.hashCode;
  }
}

class LocationConfig {
  final String name;
  final String upstream;
  final List<String> prefixes;
  final List<String> rewrites;
  final List<String> respHeaders;
  final List<String> reqHeaders;
  final List<String> hosts;
  final String proxyTimeout;
  LocationConfig({
    this.name,
    this.upstream,
    this.prefixes,
    this.rewrites,
    this.respHeaders,
    this.reqHeaders,
    this.hosts,
    this.proxyTimeout,
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
    );
  }

  String toJson() => json.encode(toMap());

  factory LocationConfig.fromJson(String source) =>
      LocationConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'LocationConfig(name: $name, upstream: $upstream, prefixes: $prefixes, rewrites: $rewrites, respHeaders: $respHeaders, reqHeaders: $reqHeaders, hosts: $hosts, proxyTimeout: $proxyTimeout)';
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
        o.proxyTimeout == proxyTimeout;
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
        proxyTimeout.hashCode;
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
  ServerConfig({
    this.logFormat,
    this.addr,
    this.locations,
    this.cache,
    this.compress,
    this.compressMinLength,
    this.compressContentTypeFilter,
  });

  ServerConfig copyWith({
    String logFormat,
    String addr,
    List<String> locations,
    String cache,
    String compress,
    String compressMinLength,
    String compressContentTypeFilter,
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
    );
  }

  String toJson() => json.encode(toMap());

  factory ServerConfig.fromJson(String source) =>
      ServerConfig.fromMap(json.decode(source));

  @override
  String toString() {
    return 'ServerConfig(logFormat: $logFormat, addr: $addr, locations: $locations, cache: $cache, compress: $compress, compressMinLength: $compressMinLength, compressContentTypeFilter: $compressContentTypeFilter)';
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
        o.compressContentTypeFilter == compressContentTypeFilter;
  }

  @override
  int get hashCode {
    return logFormat.hashCode ^
        addr.hashCode ^
        locations.hashCode ^
        cache.hashCode ^
        compress.hashCode ^
        compressMinLength.hashCode ^
        compressContentTypeFilter.hashCode;
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