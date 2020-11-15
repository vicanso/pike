///
/// 登录页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:web/config/application.dart';

import '../bloc/bloc.dart';
import '../widget/button.dart';
import '../widget/error_message.dart';

@immutable
class LoginPage extends StatefulWidget {
  const LoginPage({
    Key key,
  }) : super(key: key);
  @override
  _LoginPageState createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final GlobalKey _formKey = GlobalKey<FormState>();
  final TextEditingController _unameController = TextEditingController();
  final TextEditingController _pwdController = TextEditingController();

  UserBloc _userBloc;

  @override
  void initState() {
    super.initState();
    _userBloc = BlocProvider.of<UserBloc>(context);
  }

  void _login() {
    _userBloc.add(UserLogin(
      account: _unameController.text,
      password: _pwdController.text,
    ));
  }

  Widget _renderBody(UserState state) {
    final formItems = <Widget>[];
    // 图标
    formItems.add(Container(
      width: double.infinity,
      margin: EdgeInsets.only(
        top: 10,
        bottom: 10,
      ),
      child: Column(
        children: [
          Icon(
            Icons.hd,
            size: 100,
          ),
        ],
      ),
    ));

    // 账户输入
    formItems.add(TextFormField(
      autofocus: true,
      controller: _unameController,
      decoration: InputDecoration(
        labelText: 'Account',
        hintText: 'Please input the account',
        icon: Icon(Icons.person),
      ),
      // 校验用户名
      validator: (v) => v.trim().isNotEmpty ? null : 'account can not be nil',
    ));

    // 密码输入
    formItems.add(TextFormField(
      controller: _pwdController,
      decoration: InputDecoration(
        labelText: 'Password',
        hintText: 'Please input the password',
        icon: Icon(Icons.lock),
      ),
      obscureText: true,
      //校验密码
      validator: (v) => v.trim().isNotEmpty ? null : 'password can not be nil',
    ));
    if (state is UserErrorState) {
      // 提示
      formItems.add(Container(
        margin: EdgeInsets.only(
          top: 2 * Application.defaultPadding,
        ),
        child: XErrorTips(
          message: state.message,
        ),
      ));
    }

    var isProcessing = false;
    if (state is UserMeState && state.isProcessing) {
      isProcessing = true;
    }

    // 登录、注册按钮
    formItems.add(Padding(
      padding: const EdgeInsets.only(top: 28.0),
      child: XFullButton(
        padding: EdgeInsets.all(20.0),
        margin: EdgeInsets.all(0),
        text: Text(isProcessing ? "Login..." : "Login"),
        onPressed: () {
          // 避免多次点击登录
          if (isProcessing) {
            return;
          }
          // 校验数据
          if ((_formKey.currentState as FormState).validate()) {
            _login();
            return;
          }
          //   if (_isLoginMode) {
          //     _login();
          //   } else {
          //     _register();
          //   }
          // }
        },
      ),
    ));

    return SingleChildScrollView(
      child: Padding(
        padding: EdgeInsets.only(
          left: 20,
          right: 20,
        ),
        child: Form(
          key: _formKey, //设置globalKey，用于后面获取FormState
          child: Column(
            children: formItems,
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<UserBloc, UserState>(builder: (context, state) {
        if (state is UserMeState && state.isLogin) {
          Future.delayed(
            Duration(
              milliseconds: 100,
            ),
            () => Application.routes.goBack(context),
          );
        }
        return Scaffold(
          appBar: AppBar(
            title: Text("Login"),
          ),
          body: _renderBody(state),
        );
      });
}
