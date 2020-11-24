///
/// Admin配置页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../helper/util.dart';
import '../model/config.dart';
import '../widget/button.dart';
import '../widget/error_message.dart';

@immutable
class AdminPage extends StatefulWidget {
  const AdminPage({
    Key key,
  }) : super(key: key);
  @override
  _AdminPageState createState() => _AdminPageState();
}

class _AdminPageState extends State<AdminPage> {
  final _formKey = GlobalKey<FormState>();

  final _accountCtrl = TextEditingController();
  final _passCtrl = TextEditingController();

  ConfigBloc _configBloc;

  @override
  void initState() {
    super.initState();
    _configBloc = BlocProvider.of<ConfigBloc>(context);
  }

  // _fillEditor 填充编辑器
  void _fillEditor(AdminConfig admin) {
    if (admin == null) {
      return;
    }
    _accountCtrl.value = TextEditingValue(text: admin.user ?? '');
    _passCtrl.clear();
  }

  // _updateAdmin 更新admin
  void _updateAdmin(ConfigCurrentState state) {
    final adminConfig = AdminConfig(
      user: _accountCtrl.text.trim(),
      password: hashPassword(_passCtrl.text.trim()),
    );
    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        admin: adminConfig,
      ),
    ));
  }

  // _renderEditor 渲染编辑器
  Widget _renderEditor(ConfigCurrentState state) {
    final formItems = <Widget>[];
    if (state.config.admin != null) {
      _fillEditor(state.config.admin);
    }
    // 用户
    formItems.add(TextFormField(
      autofocus: true,
      controller: _accountCtrl,
      // initialValue: state.config.admin?.user ,
      decoration: InputDecoration(
        labelText: 'Account',
        hintText: 'Please input the account',
      ),
      validator: (v) => v.trim().isNotEmpty ? null : 'account can not be null',
    ));

    // 密码
    formItems.add(TextFormField(
      controller: _passCtrl,
      decoration: InputDecoration(
        labelText: 'Password',
        hintText: 'Please input the password',
      ),
      obscureText: true,
      validator: (v) => v.trim().isNotEmpty ? null : 'password can not be null',
    ));

    return Container(
      margin: EdgeInsets.only(
        top: 3 * Application.defaultPadding,
      ),
      child: Form(
        key: _formKey, //设置globalKey，用于后面获取FormState
        child: Column(
          children: formItems,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<ConfigBloc, ConfigState>(builder: (context, state) {
        if (state is ConfigErrorState) {
          return XErrorMessage(
            message: state.message,
            title: 'Get admin config fail',
          );
        }
        final currentConfig = state as ConfigCurrentState;
        var btnText = 'Save Admin';
        if (currentConfig.isProcessing) {
          btnText = 'Processing...';
        }
        return SingleChildScrollView(
          child: Container(
            margin: EdgeInsets.all(3 * Application.defaultPadding),
            child: Column(
              children: [
                _renderEditor(currentConfig),
                XFullButton(
                  margin: EdgeInsets.only(
                    top: 3 * Application.defaultPadding,
                  ),
                  onPressed: () {
                    if (currentConfig.isProcessing) {
                      return;
                    }
                    if (_formKey.currentState.validate()) {
                      _updateAdmin(currentConfig);
                    }
                  },
                  text: Text(btnText),
                ),
              ],
            ),
          ),
        );
      });
}
