import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:fluro/fluro.dart';

import './bloc/bloc.dart';
import './config/application.dart';
import './helper/util.dart';
import './router/routes.dart';
import './service/store.dart' as store;

Future main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await store.init();
  } finally {
    runZoned(() {
      runApp(MyApp());
    }, onError: (err, trace) {
      showErrorMessage(err.toString());
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
        // 应用信息bloc
        BlocProvider<ApplicationInfoBloc>(
          create: (context) => ApplicationInfoBloc(),
        ),
        // 缓存bloc
        BlocProvider<CacheBloc>(
          create: (context) => CacheBloc(),
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
