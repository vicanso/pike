///
/// 导航相关事件
///

abstract class NavigationEvent {}

// 拉取首页导航
class MainNavigationFetch extends NavigationEvent {}

// 首页导航当前index修改
class MainNavigationCurrentIndexChange extends NavigationEvent {
  final int index;
  MainNavigationCurrentIndexChange({
    this.index,
  });
}
