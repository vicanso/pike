import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:fluro/fluro.dart';
import 'dart:async';

import './bloc/bloc.dart';
import './config/application.dart';
import './service/store.dart' as store;
import './router/routes.dart';

Future main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await store.init();
  } finally {
    runZoned(() {
      runApp(MyApp());
    }, onError: (err, trace) {
      // TODO 添加出错日志
    });
  }
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final router = FluroRouter();
    final routes = Routes(
      router: router,
    )..init();
    Application.routes = routes;
    return MultiBlocProvider(
      providers: [
        // 主导航bloc
        BlocProvider<MainNavigationBloc>(
          create: (context) => MainNavigationBloc()..add(MainNavigationFetch()),
        ),
        // 用户信息bloc
        BlocProvider<UserBloc>(
          create: (context) => UserBloc()..add(UserMeFetch()),
        ),
      ],
      child: MaterialApp(
        theme: ThemeData(
          primaryColor: Application.primaryColor,
          visualDensity: VisualDensity.adaptivePlatformDensity,
        ),
        // 页面内容由route来控制展示
        onGenerateRoute: router.generator,
      ),
    );
  }
}
