///
/// upstream配置页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../helper/util.dart';
import '../model/config.dart';
import '../widget/error_message.dart';

@immutable
class UpstreamPage extends StatefulWidget {
  const UpstreamPage({
    Key key,
  }) : super(key: key);
  @override
  _UpstreamPageState createState() => _UpstreamPageState();
}

class _UpstreamPageState extends State<UpstreamPage> {
  final GlobalKey _formKey = GlobalKey<FormState>();

  String _mode = '';
  final _editMode = 'eidt';
  final _updateMode = 'update';

  ConfigBloc _configBloc;

  @override
  void initState() {
    super.initState();
    _configBloc = BlocProvider.of<ConfigBloc>(context);
  }

  bool get _isEditting => _mode.isNotEmpty;

  bool get _isUpdateding => _mode == _updateMode;

  // _reset 重置表单所有元素
  void _reset() {}

  // _fillTextEditor 填充编辑数据
  void _fillTextEditor(UpstreamConfig element) {}

  // _createRowItem 生成表单元素
  Widget _createRowItem(String text) => Padding(
        padding: EdgeInsets.only(
          top: Application.defaultPadding,
          bottom: Application.defaultPadding,
        ),
        child: Text(
          text ?? '--',
          textAlign: TextAlign.center,
        ),
      );

  void _deleteUpstream(ConfigCurrentState state, String name) {
    // 校验该upstream是否被其它配置使用
    if (!state.config.validateForDelete('upstream', name)) {
      showErrorMessage('$name is used, it cant not be deleted');
      return;
    }
    final upstreamList = <UpstreamConfig>[];
    state.config.upstreams?.forEach((element) {
      if (element.name != name) {
        upstreamList.add(element);
      }
    });
    // 更新配置
    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        upstreams: upstreamList,
      ),
    ));
  }

  // _renderServerList 渲染服务器列表
  Widget _renderServerList(List<UpstreamServerConfig> servers) {
    final items = servers?.map((element) {
      var addr = element.addr;
      if (element.backup != null && element.backup) {
        addr += ' (backup)';
      }
      return Text(addr);
    })?.toList();
    return Padding(
      padding: EdgeInsets.only(
        top: Application.defaultPadding,
        bottom: Application.defaultPadding,
      ),
      child: Column(
        children: items,
      ),
    );
  }

  // _renderUpstreamList 渲染upstream列表
  Widget _renderUpstreamList(ConfigCurrentState state) {
    // 表头
    final rows = <TableRow>[
      TableRow(
        children: [
          _createRowItem('Name'),
          _createRowItem('Health Check'),
          _createRowItem('Policy'),
          _createRowItem('Enable H2C'),
          _createRowItem('Accept Encoding'),
          _createRowItem('Servers'),
          _createRowItem('Remark'),
          _createRowItem('Operations'),
        ],
      ),
    ];
    state.config.upstreams?.forEach((element) {
      var enableH2C = 'off';
      if (element.enableH2C != null && element.enableH2C) {
        enableH2C = 'on';
      }
      rows.add(TableRow(
        children: [
          _createRowItem(element.name),
          _createRowItem(element.healthCheck ?? ''),
          _createRowItem(element.policy ?? ''),
          _createRowItem(enableH2C),
          _createRowItem(element.acceptEncoding ?? ''),
          _renderServerList(element.servers),
          _createRowItem(element.remark),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              TextButton(
                onPressed: () {
                  // 重置当前数据，并将需要更新的配置填充
                  _reset();
                  _fillTextEditor(element);

                  setState(() {
                    _mode = _updateMode;
                  });
                },
                child: Text('Update'),
              ),
              TextButton(
                onPressed: () {
                  _deleteUpstream(state, element.name);
                },
                child: Text('Delete'),
              ),
            ],
          ),
        ],
      ));
    });
    return Table(
      columnWidths: {
        // 指定表格列宽
        7: FixedColumnWidth(150),
      },
      border: TableBorder.all(
        color: Application.primaryColor.withAlpha(60),
      ),
      children: rows,
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
        print(currentConfig);
        return SingleChildScrollView(
          child: Container(
            margin: EdgeInsets.all(3 * Application.defaultPadding),
            child: Column(
              children: [
                _renderUpstreamList(currentConfig),
              ],
            ),
          ),
        );
      });
}
