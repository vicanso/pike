///
/// YAML配置信息
///
import 'dart:convert';
import 'dart:html' as html;
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../model/config.dart';
import '../widget/card.dart';
import '../widget/error_message.dart';

@immutable
class YAMLConfigPage extends StatefulWidget {
  const YAMLConfigPage({
    Key key,
  }) : super(key: key);
  @override
  _YAMLConfigPageState createState() => _YAMLConfigPageState();
}

const editMode = 'edit';

class _YAMLConfigPageState extends State<YAMLConfigPage> {
  String _mode = '';
  final _yamlCtrl = TextEditingController();

  ConfigBloc _configBloc;

  @override
  void initState() {
    super.initState();
    _configBloc = BlocProvider.of<ConfigBloc>(context);
  }

  Widget _renderContent(String yaml) {
    if (_mode != editMode) {
      return Text(
        yaml ?? '-- No Content --',
        style: TextStyle(
          height: 1.5,
        ),
      );
    }

    return TextFormField(
      controller: _yamlCtrl,
      minLines: 20,
      maxLines: 30,
      decoration: InputDecoration(
        labelText: 'YAML Config',
        hintText: 'Please input the content of yaml',
      ),
    );
  }

  // _render 渲染yaml的配置
  Widget _render(ConfigCurrentState state) {
    final yaml = state.config.yaml;
    final padding = Container(
      width: Application.defaultPadding,
    );
    final actions = <Widget>[];
    // 非编辑模式
    if (_mode != editMode) {
      actions.add(
        IconButton(
          padding: EdgeInsets.all(0),
          icon: Icon(
            Icons.edit,
          ),
          onPressed: () {
            setState(() {
              _mode = editMode;
            });
          },
        ),
      );
      actions.add(padding);
    } else {
      // 编辑模式之后
      _yamlCtrl.value = TextEditingValue(text: yaml ?? '');

      // 添加action按钮
      actions.add(
        IconButton(
          padding: EdgeInsets.all(0),
          icon: Icon(
            Icons.cancel,
          ),
          onPressed: () {
            setState(() {
              _mode = '';
            });
          },
        ),
      );
      actions.add(padding);
      actions.add(
        IconButton(
          padding: EdgeInsets.all(0),
          icon: Icon(
            Icons.save,
          ),
          onPressed: () {
            setState(() {
              _mode = '';
              _configBloc.add(ConfigUpdate(
                config: Config(
                  yaml: _yamlCtrl.value.text,
                ),
              ));
            });
          },
        ),
      );
    }
    // 如果有配置文件，则添加下载按钮
    if (_mode != editMode && yaml != null && yaml.isNotEmpty) {
      actions.add(IconButton(
        padding: EdgeInsets.all(0),
        constraints: BoxConstraints(),
        icon: Icon(
          Icons.cloud_download,
        ),
        onPressed: () {
          final bytes = utf8.encode(yaml);
          final blob = html.Blob([bytes]);
          final url = html.Url.createObjectUrlFromBlob(blob);
          final anchor = html.document.createElement('a') as html.AnchorElement
            ..href = url
            ..style.display = 'none'
            ..download = 'pike.yaml';
          html.document.body.children.add(anchor);
          // download
          anchor.click();
          // cleanup
          html.document.body.children.remove(anchor);
          html.Url.revokeObjectUrl(url);
        },
      ));
    }

    return XCard(
      'Config',
      _renderContent(yaml),
      actions: actions,
    );
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<ConfigBloc, ConfigState>(builder: (context, state) {
        if (state is ConfigErrorState) {
          return XErrorMessage(
            message: state.message,
            title: 'Get upstream config fail',
          );
        }
        final currentConfig = state as ConfigCurrentState;
        return _render(currentConfig);
      });
}
