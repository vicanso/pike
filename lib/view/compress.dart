///
/// 压缩配置页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../helper/util.dart';
import '../model/config.dart';
import '../widget/button.dart';
import '../widget/error_message.dart';
import '../widget/table.dart';

@immutable
class CompressPage extends StatefulWidget {
  const CompressPage({
    Key key,
  }) : super(key: key);
  @override
  _CompressPageState createState() => _CompressPageState();
}

class _CompressPageState extends State<CompressPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtrl = TextEditingController();
  final _gzipCtrl = TextEditingController();
  final _brCtrl = TextEditingController();
  final _remarkCtrl = TextEditingController();
  String _mode = '';
  final _editMode = 'eidt';
  final _updateMode = 'update';
  final _gzipName = 'gzip';
  final _brName = 'br';

  ConfigBloc _configBloc;

  @override
  void initState() {
    super.initState();
    _configBloc = BlocProvider.of<ConfigBloc>(context);
  }

  bool get _isEditting => _mode.isNotEmpty;

  bool get _isUpdating => _mode == _updateMode;

  // _getLevelDesc 获取压缩级别的描述
  String _getLevelDesc(Map<String, int> levels, String name) {
    final value = levels[name];
    if (value == null) {
      return '--';
    }
    return value.toString();
  }

  // _getLevel 获取压缩级别
  String _getLevel(Map<String, int> levels, String name) {
    final value = levels[name];
    if (value == null) {
      return '';
    }
    return value.toString();
  }

  // _reset 重置表单所有元素
  void _reset() {
    _nameCtrl.clear();
    _gzipCtrl.clear();
    _brCtrl.clear();
    _remarkCtrl.clear();
  }

  // _fillEditor 填充编辑数据
  void _fillEditor(CompressConfig element) {
    _nameCtrl.value = TextEditingValue(text: element.name ?? '');
    _gzipCtrl.value = TextEditingValue(
        text: _getLevel(
      element.levels,
      _gzipName,
    ));
    _brCtrl.value = TextEditingValue(
        text: _getLevel(
      element.levels,
      _brName,
    ));
    _remarkCtrl.value = TextEditingValue(
      text: element.remark ?? '',
    );
  }

  // _deleteCompress 删除压缩
  void _deleteCompress(ConfigCurrentState state, String name) {
    // 校验该压缩是否被其它配置使用
    if (!state.config.validateForDelete('compress', name)) {
      showErrorMessage('$name is used, it can not be deleted');
      return;
    }
    final compressList = <CompressConfig>[];

    state.config.compresses?.forEach((element) {
      if (element.name != name) {
        compressList.add(element);
      }
    });
    // 更新配置
    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        compresses: compressList,
      ),
    ));
  }

  // _renderCompressList 渲染当前压缩服务列表
  Widget _renderCompressList(ConfigCurrentState state) {
    // 表格内容，压缩服务的相关配置
    final contents = state.config.compresses
        ?.map((e) => [
              e.name,
              _getLevelDesc(e.levels, _gzipName),
              _getLevelDesc(e.levels, _brName),
              e.remark,
            ])
        ?.toList();
    final doUpdate = (int index) {
      final element = state.config.compresses.elementAt(index);
      // 重置当前数据，并将需要更新的配置填充
      _reset();
      _fillEditor(element);

      setState(() {
        _mode = _updateMode;
      });
    };
    final doDelete = (int index) {
      final element = state.config.compresses.elementAt(index);
      _deleteCompress(state, element.name);
    };
    final gzipHeaderName = _gzipName[0].toUpperCase() + _gzipName.substring(1);
    final brHeaderName = _brName[0].toLowerCase() + _brName.substring(1);
    return XConfigTable(
      headers: [
        'Name',
        gzipHeaderName,
        brHeaderName,
        'Remark',
      ],
      contents: contents,
      onUpdate: doUpdate,
      onDelete: doDelete,
      columnWidths: <String, double>{
        gzipHeaderName: 80,
        brHeaderName: 80,
      },
    );
  }

  // _renderEditor 渲染编辑表单
  Widget _renderEditor() {
    if (!_isEditting) {
      return Container();
    }
    final fn =
        createNumberValidator('compress level should be number and gt 0');
    final formItems = <Widget>[];
    // 名称
    formItems.add(TextFormField(
      autofocus: true,
      readOnly: _isUpdating,
      controller: _nameCtrl,
      decoration: InputDecoration(
        labelText: 'Name',
        hintText: 'Please input the name of compress',
      ),
      validator: (v) => v.trim().isNotEmpty ? null : 'name can not be null',
    ));

    // gzip 压缩级别
    formItems.add(TextFormField(
      controller: _gzipCtrl,
      decoration: InputDecoration(
        labelText: 'Gzip Level',
        hintText: 'Please input the compress level of gzip(1-9)',
      ),
      validator: fn,
    ));

    // br 压缩级别
    formItems.add(TextFormField(
      controller: _brCtrl,
      decoration: InputDecoration(
        labelText: 'Br Level',
        hintText: 'Please input the compress level of br(1-11)',
      ),
      validator: fn,
    ));

    // remark
    formItems.add(TextFormField(
      controller: _remarkCtrl,
      minLines: 3,
      maxLines: 3,
      decoration: InputDecoration(
        labelText: 'Remark',
        hintText: 'Please input the remark for compress',
      ),
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

  // _addCompress 添加压缩服务，如果添加的服务名称与当前服务相同，则替换
  void _addCompress(ConfigCurrentState state) {
    final name = _nameCtrl.text?.trim();
    final levels = <String, int>{
      'gzip': int.parse(_gzipCtrl.text),
      'br': int.parse(_brCtrl.text),
    };
    final compressConfig = CompressConfig(
      name: name,
      levels: levels,
      remark: _remarkCtrl.text?.trim(),
    );
    final compressList = <CompressConfig>[];
    state.config.compresses?.forEach((element) {
      if (element.name != name) {
        compressList.add(element);
      }
    });
    compressList.add(compressConfig);

    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        compresses: compressList,
      ),
    ));
    // 重置当前模式
    setState(() {
      _mode = '';
    });
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<ConfigBloc, ConfigState>(builder: (context, state) {
        if (state is ConfigErrorState) {
          return XErrorMessage(
            message: state.message,
            title: 'Get compress config fail',
          );
        }
        final currentConfig = state as ConfigCurrentState;
        var btnText = _isEditting ? 'Save Compress' : 'Add Compress';
        if (currentConfig.isProcessing) {
          btnText = 'Processing...';
        }
        return SingleChildScrollView(
          child: Container(
            margin: EdgeInsets.all(3 * Application.defaultPadding),
            child: Column(
              children: [
                _renderCompressList(currentConfig),
                _renderEditor(),
                XFullButton(
                  margin: EdgeInsets.only(
                    top: 3 * Application.defaultPadding,
                  ),
                  onPressed: () {
                    if (currentConfig.isProcessing) {
                      return;
                    }
                    // 如果是编辑模式，则是添加或更新
                    if (_isEditting) {
                      if (_formKey.currentState.validate()) {
                        _addCompress(currentConfig);
                      }
                      return;
                    }
                    // 重置所有数据，设置为编辑模式
                    _reset();
                    setState(() {
                      _mode = _editMode;
                    });
                  },
                  text: Text(btnText),
                ),
              ],
            ),
          ),
        );
      });
}
