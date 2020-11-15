///
/// 应用导航条配置
///
import 'dart:convert';

import 'package:flutter/material.dart';

class NavItem {
  final String title;
  final IconData icon;
  NavItem({
    this.title,
    this.icon,
  });

  NavItem copyWith({
    String title,
    IconData icon,
  }) {
    return NavItem(
      title: title ?? this.title,
      icon: icon ?? this.icon,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'title': title,
      'icon': icon?.codePoint,
    };
  }

  factory NavItem.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return NavItem(
      title: map['title'],
      icon: IconData(map['icon'], fontFamily: 'MaterialIcons'),
    );
  }

  String toJson() => json.encode(toMap());

  factory NavItem.fromJson(String source) =>
      NavItem.fromMap(json.decode(source));

  @override
  String toString() => 'NavItem(title: $title, icon: $icon)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is NavItem && o.title == title && o.icon == icon;
  }

  @override
  int get hashCode => title.hashCode ^ icon.hashCode;
}
