///
/// 应用导航条配置
///
// ignore_for_file: argument_type_not_assignable
// ignore_for_file:  prefer_expression_function_bodies
import 'dart:convert';

import 'package:flutter/material.dart';

class NavItem {
  final String title;
  final IconData icon;
  final String name;
  NavItem({
    this.title,
    this.icon,
    this.name,
  });

  NavItem copyWith({
    String title,
    IconData icon,
    String name,
  }) {
    return NavItem(
      title: title ?? this.title,
      icon: icon ?? this.icon,
      name: name ?? this.name,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'title': title,
      'icon': icon?.codePoint,
      'name': name,
    };
  }

  factory NavItem.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return NavItem(
      title: map['title'],
      icon: IconData(map['icon'], fontFamily: 'MaterialIcons'),
      name: map['name'],
    );
  }

  String toJson() => json.encode(toMap());

  factory NavItem.fromJson(String source) =>
      NavItem.fromMap(json.decode(source));

  @override
  String toString() => 'NavItem(title: $title, icon: $icon, name: $name)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is NavItem && o.title == title && o.icon == icon && o.name == name;
  }

  @override
  int get hashCode => title.hashCode ^ icon.hashCode ^ name.hashCode;
}
