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
  const MainNavigationSuccess({
    this.navs,
    this.currentIndex,
  });

  factory MainNavigationSuccess.newDefault() => MainNavigationSuccess(
        currentIndex: 0,
        navs: [
          NavItem(
            title: 'Home',
            icon: Icons.sensor_window,
            name: 'home',
          ),
          NavItem(
            title: 'Compress',
            icon: Icons.filter_frames,
            name: 'compress',
          ),
          NavItem(
            title: 'Cache',
            icon: Icons.storage,
            name: 'cache',
          ),
          NavItem(
            title: 'Upstram',
            icon: Icons.dvr,
            name: 'upstream',
          ),
          NavItem(
            title: 'Location',
            icon: Icons.alt_route,
            name: 'location',
          ),
          NavItem(
            title: 'Server',
            icon: Icons.widgets,
            name: 'server',
          ),
          NavItem(
            title: 'Admin',
            icon: Icons.admin_panel_settings,
            name: 'admin',
          ),
          NavItem(
            title: 'Caches',
            icon: Icons.format_list_bulleted,
            name: 'caches',
          ),
        ],
      );

  @override
  List<Object> get props => [navs, currentIndex];

  MainNavigationSuccess copyWith({
    List<NavItem> navs,
    int currentIndex,
  }) =>
      MainNavigationSuccess(
        navs: navs ?? this.navs,
        currentIndex: currentIndex ?? this.currentIndex,
      );

  @override
  String toString() =>
      'MainNavigationSuccess(navs: $navs, currentIndex: $currentIndex)';
}
