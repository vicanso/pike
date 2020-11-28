///
/// 应用信息
///
// ignore_for_file: argument_type_not_assignable
// ignore_for_file:  prefer_expression_function_bodies
import 'dart:convert';

import 'package:flutter/foundation.dart';

void fillEmptyList(Map<String, dynamic> m) {
  m['processing'] ??= <String, int>{};
}

class ApplicationInfo {
  final String goarch;
  final String goos;
  final String goVersion;
  final String version;
  final String buildedAt;
  final String commitID;
  final String uptime;
  final int goMaxProcs;
  final int routineCount;
  final Map<String, int> processing;
  ApplicationInfo({
    this.goarch,
    this.goos,
    this.goVersion,
    this.version,
    this.buildedAt,
    this.commitID,
    this.uptime,
    this.goMaxProcs,
    this.routineCount,
    this.processing,
  });

  ApplicationInfo copyWith({
    String goarch,
    String goos,
    String goVersion,
    String version,
    String buildedAt,
    String commitID,
    String uptime,
    int goMaxProcs,
    int routineCount,
    Map<String, int> processing,
  }) {
    return ApplicationInfo(
      goarch: goarch ?? this.goarch,
      goos: goos ?? this.goos,
      goVersion: goVersion ?? this.goVersion,
      version: version ?? this.version,
      buildedAt: buildedAt ?? this.buildedAt,
      commitID: commitID ?? this.commitID,
      uptime: uptime ?? this.uptime,
      goMaxProcs: goMaxProcs ?? this.goMaxProcs,
      routineCount: routineCount ?? this.routineCount,
      processing: processing ?? this.processing,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'goarch': goarch,
      'goos': goos,
      'goVersion': goVersion,
      'version': version,
      'buildedAt': buildedAt,
      'commitID': commitID,
      'uptime': uptime,
      'goMaxProcs': goMaxProcs,
      'routineCount': routineCount,
      'processing': processing,
    };
  }

  factory ApplicationInfo.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    fillEmptyList(map);

    return ApplicationInfo(
      goarch: map['goarch'],
      goos: map['goos'],
      goVersion: map['goVersion'],
      version: map['version'],
      buildedAt: map['buildedAt'],
      commitID: map['commitID'],
      uptime: map['uptime'],
      goMaxProcs: map['goMaxProcs'],
      routineCount: map['routineCount'],
      processing: Map<String, int>.from(map['processing']),
    );
  }

  String toJson() => json.encode(toMap());

  factory ApplicationInfo.fromJson(String source) =>
      ApplicationInfo.fromMap(json.decode(source));

  @override
  String toString() {
    return 'ApplicationInfo(goarch: $goarch, goos: $goos, goVersion: $goVersion, version: $version, buildedAt: $buildedAt, commitID: $commitID, uptime: $uptime, goMaxProcs: $goMaxProcs, routineCount: $routineCount, processing: $processing)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is ApplicationInfo &&
        o.goarch == goarch &&
        o.goos == goos &&
        o.goVersion == goVersion &&
        o.version == version &&
        o.buildedAt == buildedAt &&
        o.commitID == commitID &&
        o.uptime == uptime &&
        o.goMaxProcs == goMaxProcs &&
        o.routineCount == routineCount &&
        mapEquals(o.processing, processing);
  }

  @override
  int get hashCode {
    return goarch.hashCode ^
        goos.hashCode ^
        goVersion.hashCode ^
        version.hashCode ^
        buildedAt.hashCode ^
        commitID.hashCode ^
        uptime.hashCode ^
        goMaxProcs.hashCode ^
        routineCount.hashCode ^
        processing.hashCode;
  }
}
