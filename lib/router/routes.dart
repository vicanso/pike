///
/// 路由配置
///
import 'package:flutter/material.dart';
import 'package:fluro/fluro.dart';

import '../view/home.dart';
import '../view/login.dart';

class Routes {
  final FluroRouter router;
  String _currentRoute;

  final String _root = '/';
  final String _login = '/login';

  Routes({
    @required this.router,
  });

  String get currentRoute => _currentRoute;

  Handler _createHandler(String route, HandlerFunc fn) => Handler(
        handlerFunc: (BuildContext context, Map<String, List<String>> params) {
          _currentRoute = route;
          return fn(context, params);
        },
      );

  void _define(String route, HandlerFunc fn) {
    router.define(
      route,
      handler: _createHandler(route, fn),
    );
  }

  void init() {
    // 首页
    _define(_root, (context, parameters) => HomePage());
    // 登录页面
    _define(_login, (context, parameters) => LoginPage());
  }

  void _navigateTo(context, path) {
    router.navigateTo(
      context,
      path,
      transition: TransitionType.fadeIn,
    );
  }

  // 跳转至首页
  void goToHome(BuildContext context) {
    _navigateTo(context, _root);
  }

  // 跳转至登录页
  void goToLogin(BuildContext context) {
    _navigateTo(context, _login);
  }

  // goBack 返回
  void goBack(BuildContext context) {
    router.pop(context);
  }
}
