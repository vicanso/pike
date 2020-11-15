///
/// 首页导航相关的state
///
import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';

import '../model/navigation.dart';

abstract class NavigationState extends Equatable {
  const NavigationState();
  @override
  List<Object> get props => [];
}

// 初始化首页导航
class MainNavigationInitial extends NavigationState {}

// 获取首页导航失败
class MainNavigationFailure extends NavigationState {}

class MainNavigationSuccess extends NavigationState {
  final List<NavItem> navs;
  final int currentIndex;
  MainNavigationSuccess({
    this.navs,
    this.currentIndex,
  });

  factory MainNavigationSuccess.newDefault() => MainNavigationSuccess(
        currentIndex: 0,
        navs: [
          NavItem(title: 'Home', icon: Icons.home),
          NavItem(title: 'Compress', icon: Icons.home),
          NavItem(title: 'Cache', icon: Icons.category),
          NavItem(title: 'Upstram', icon: Icons.shopping_cart),
          NavItem(title: 'Location', icon: Icons.perm_identity),
          NavItem(title: 'Server', icon: Icons.perm_identity),
          NavItem(title: 'Admin', icon: Icons.perm_identity),
        ],
      );

  @override
  List<Object> get props => [navs, currentIndex];

  MainNavigationSuccess copyWith({
    List<NavItem> navs,
    int currentIndex,
  }) {
    return MainNavigationSuccess(
      navs: navs ?? this.navs,
      currentIndex: currentIndex ?? this.currentIndex,
    );
  }

  @override
  String toString() =>
      'MainNavigationSuccess(navs: $navs, currentIndex: $currentIndex)';
}
