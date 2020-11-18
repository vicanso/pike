///
/// 首页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../widget/error_message.dart';
import './cache.dart';
import './compress.dart';
import './upstream.dart';

@immutable
class HomePage extends StatefulWidget {
  final int currentIndex;
  final String recommender;
  const HomePage({
    this.currentIndex,
    this.recommender,
    Key key,
  }) : super(key: key);
  @override
  _HomePageState createState() => _HomePageState();
}

class _HomePageState extends State<HomePage>
    with SingleTickerProviderStateMixin {
  bool _isLogin = false;
  bool _isFetchingUserInfo = true;
  TabController _tabController;
  int _currentIndex = 0;
  ConfigBloc _configBloc;

  PreferredSizeWidget _renderAppBar(MainNavigationSuccess state) {
    // 如果未登录，则不需要展示
    if (!_isLogin) {
      return null;
    }
    final tabs = state.navs
        .map((e) => Tab(
              iconMargin: EdgeInsets.only(
                bottom: 0.5 * Application.defaultPadding,
              ),
              icon: Icon(
                e.icon,
              ),
              text: e.title,
            ))
        .toList();
    _tabController ??= TabController(
      length: tabs.length,
      vsync: this,
    )..addListener(_handleNavigationSelect);
    return PreferredSize(
      preferredSize: Size.fromHeight(Application.navigationHeight),
      child: Container(
        padding: EdgeInsets.only(
          top: 0.5 * Application.defaultPadding,
        ),
        color: Theme.of(context).primaryColor,
        child: TabBar(
          labelColor: Application.fontColorOfPrimaryColor,
          // isScrollable: true,
          indicatorColor: Application.blueColor,
          tabs: tabs,
          controller: _tabController,
        ),
      ),
    );
  }

  void _handleNavigationSelect() {
    if (!_tabController.indexIsChanging) {
      return;
    }
    setState(() {
      _currentIndex = _tabController.index;
    });
  }

  Widget _renderYAMLConfig(ConfigCurrentState state) {
    final exp = RegExp(r'(password:[\S\s]+?\n)');
    final yaml = state.config.yaml.replaceFirst(exp, 'password: ***\n');
    return SingleChildScrollView(
      child: Container(
        margin: EdgeInsets.all(2 * Application.defaultPadding),
        child: Text(yaml),
      ),
    );
  }

  Widget _renderConfig(ConfigState state) {
    if (state is ConfigErrorState) {
      return XErrorMessage(
        message: state.message,
        title: 'Fetch config fail',
      );
    }
    final configState = state as ConfigCurrentState;
    // 如果未加载到配置，展示拉取中
    if (configState.config == null) {
      return Center(
        child: Text('Fetching config...'),
      );
    }
    switch (_currentIndex) {
      case 0:
        // 渲染yaml详细配置
        return _renderYAMLConfig(configState);
        break;
      case 1:
        // 压缩配置
        return CompressPage();
        break;
      case 2:
        // 缓存配置
        return CachePage();
        break;
      case 3:
        // upstream配置
        return UpstreamPage();
        break;
      default:
    }
    return Container();
  }

  Widget _renderBody() {
    if (_isFetchingUserInfo) {
      return Center(
        child: Text('Fetching user informations...'),
      );
    }
    return BlocProvider(
      create: (context) {
        // 首次初始化时触发fetch
        if (_configBloc == null) {
          _configBloc = ConfigBloc();
          _configBloc.add(ConfigFetch());
        }
        return _configBloc;
      },
      child: BlocBuilder<ConfigBloc, ConfigState>(
        builder: (context, state) => _renderConfig(state),
      ),
    );
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<MainNavigationBloc, NavigationState>(
          builder: (context, state) {
        if (state is! MainNavigationSuccess) {
          return Scaffold(
            body: Center(
              child: Text('Loading...'),
            ),
          );
        }
        final successState = state as MainNavigationSuccess;
        return BlocListener<UserBloc, UserState>(
          listener: (context, state) {
            if (state is UserMeState) {
              if (!state.isProcessing) {
                // 如果未登录，则跳转登录
                if (!state.isLogin) {
                  Application.routes.goToLogin(context);
                  return;
                }
                setState(() {
                  _isFetchingUserInfo = false;
                  _isLogin = state.isLogin;
                });
              }
            }
          },
          child: Scaffold(
            appBar: _renderAppBar(successState),
            body: _renderBody(),
          ),
        );
      });
}
