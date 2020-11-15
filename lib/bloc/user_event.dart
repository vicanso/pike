///
/// 用户事件
///

abstract class UserEvent {}

// 用户信息重置
class UserMeReset extends UserEvent {}

// 拉取用户信息
class UserMeFetch extends UserEvent {}

// 用户登录
class UserLogin extends UserEvent {
  final String account;
  final String password;
  UserLogin({
    this.account,
    this.password,
  });
}

// 退出登录
class UserLogout extends UserEvent {}
