///
/// 导航bloc
///
import 'package:bloc/bloc.dart';

import './navigation_event.dart';
import './navigation_state.dart';

// 首页导航的bloc
class MainNavigationBloc extends Bloc<NavigationEvent, NavigationState> {
  MainNavigationBloc() : super(MainNavigationInitial());

  @override
  Stream<NavigationState> mapEventToState(NavigationEvent event) async* {
    // 拉取首页导航数据
    if (event is MainNavigationFetch) {
      try {
        // 如果是刚初始化，则拉取数据
        if (state is MainNavigationInitial) {
          yield MainNavigationSuccess.newDefault();
        }
      } on Exception catch (_) {
        yield MainNavigationFailure();
      }
      return;
    }

    // 首页导航当前index变化
    if (event is MainNavigationCurrentIndexChange) {
      final index = event.index;
      yield (state as MainNavigationSuccess).copyWith(
        currentIndex: index,
      );
      return;
    }
  }
}
