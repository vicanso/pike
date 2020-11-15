///
/// 用户相关的state
///
import 'package:equatable/equatable.dart';

import '../model/user.dart';

abstract class UserState extends Equatable {}

// 用户信息
class UserMeState extends UserState {
  final User user;
  final bool processing;
  UserMeState({
    this.user,
    this.processing,
  });

  // isProcessing 是否正在拉取用户信息
  bool get isProcessing => processing != null && processing;

  // isLogin 是否已登录
  bool get isLogin => !isProcessing && (user?.account?.isNotEmpty ?? false);

  @override
  List<Object> get props => [user, processing];

  @override
  String toString() => 'UserMeState(user: $user, processing: $processing)';
}

// 用户出错信息
class UserErrorState extends UserState {
  final String message;
  UserErrorState({
    this.message,
  });

  @override
  List<Object> get props => [message];

  @override
  String toString() => 'UserErrorState(message: $message)';
}
