///
/// 缓存配置页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../helper/util.dart';
import '../model/config.dart';
import '../widget/button.dart';
import '../widget/error_message.dart';

@immutable
class CachePage extends StatefulWidget {
  const CachePage({
    Key key,
  }) : super(key: key);
  @override
  _CachePageState createState() => _CachePageState();
}

class _CachePageState extends State<CachePage> {
  final GlobalKey _formKey = GlobalKey<FormState>();
  final TextEditingController _nameController = TextEditingController();
  final TextEditingController _sizeController = TextEditingController();
  final TextEditingController _hitForPassController = TextEditingController();
  final TextEditingController _remarkController = TextEditingController();

  ConfigBloc _configBloc;
  final numberFormat = NumberFormat('#,##0', 'en_US');

  String _mode = '';
  final _editMode = 'eidt';
  final _updateMode = 'update';

  @override
  void initState() {
    super.initState();
    _configBloc = BlocProvider.of<ConfigBloc>(context);
  }

  bool get _isEditting => _mode.isNotEmpty;

  bool get _isUpdateding => _mode == _updateMode;

  // _reset 重置表单所有元素
  void _reset() {
    _nameController.clear();
    _sizeController.clear();
    _hitForPassController.clear();
    _remarkController.clear();
  }

  // _fillTextEditor 填充编辑数据
  void _fillTextEditor(CacheConfig element) {
    _nameController.value = TextEditingValue(text: element.name);
    _sizeController.value = TextEditingValue(text: element.size.toString());
    _hitForPassController.value = TextEditingValue(text: element.hitForPass);
    _remarkController.value = TextEditingValue(text: element.remark ?? '');
  }

  // _deleteCache 删除缓存配置
  void _deleteCache(ConfigCurrentState state, String name) {
    // 校验该缓存是否被其它配置使用
    if (!state.config.validateForDelete('cache', name)) {
      showErrorMessage('$name is used, it can not be deleted');
      return;
    }
    final cacheList = <CacheConfig>[];
    state.config.caches?.forEach((element) {
      if (element.name != name) {
        cacheList.add(element);
      }
    });

    // 更新配置
    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        caches: cacheList,
      ),
    ));
  }

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

  // _renderCacheList 渲染当前缓存服务列表
  Widget _renderCacheList(ConfigCurrentState state) {
    // 表头
    final rows = <TableRow>[
      TableRow(
        children: [
          _createRowItem('Name'),
          _createRowItem('Size'),
          _createRowItem('Hit For Pass'),
          _createRowItem('Remark'),
          _createRowItem('Operations'),
        ],
      ),
    ];

    // 表格内容，缓存服务的相关配置
    state.config.caches?.forEach((element) {
      rows.add(TableRow(
        children: [
          _createRowItem(element.name),
          _createRowItem(numberFormat.format(element.size)),
          _createRowItem(element.hitForPass),
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
                  _deleteCache(state, element.name);
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
        1: FixedColumnWidth(100),
        2: FixedColumnWidth(100),
        4: FixedColumnWidth(150),
      },
      border: TableBorder.all(
        color: Application.primaryColor.withAlpha(60),
      ),
      children: rows,
    );
  }

  // _addCache 添加缓存服务，如果存在相同服务，则替换
  void _addCache(ConfigCurrentState state) {
    final cacheConfig = CacheConfig(
      name: _nameController.text?.trim(),
      size: int.parse(_sizeController.text),
      hitForPass: _hitForPassController.text?.trim(),
      remark: _remarkController.text?.trim(),
    );
    final cacheList = <CacheConfig>[];
    state.config.caches?.forEach((element) {
      if (element.name != cacheConfig.name) {
        cacheList.add(element);
      }
    });
    cacheList.add(cacheConfig);

    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        caches: cacheList,
      ),
    ));
    // 重置当前模式
    setState(() {
      _mode = '';
    });
  }

  // _renderEditor 渲染编辑表单
  Widget _renderEditor() {
    if (!_isEditting) {
      return Container();
    }
    final formItems = <Widget>[];
    // 缓存名称
    formItems.add(TextFormField(
      autofocus: true,
      readOnly: _isUpdateding,
      controller: _nameController,
      decoration: InputDecoration(
        labelText: 'Name',
        hintText: 'Please input the name of cache',
      ),
      validator: (v) => v.trim().isNotEmpty ? null : 'name can not be null',
    ));

    // 缓存大小
    formItems.add(TextFormField(
      controller: _sizeController,
      decoration: InputDecoration(
        labelText: 'Size',
        hintText: 'Please input the size of cache',
      ),
      validator:
          createNumberValidator('size of cache should be number and gt 0'),
    ));

    // hit for pass时长
    formItems.add(TextFormField(
      controller: _hitForPassController,
      decoration: InputDecoration(
        labelText: 'HitForPass',
        hintText: 'Please input the duration of hit for pass(5m, 30s)',
      ),
      validator: (value) {
        final reg = RegExp(r'^\d+[sm]$');
        if (value == null || !reg.hasMatch(value)) {
          return 'duration of hit for pass is invalid';
        }
        return null;
      },
    ));

    // remark
    formItems.add(TextFormField(
      controller: _remarkController,
      minLines: 3,
      maxLines: 3,
      decoration: InputDecoration(
        labelText: 'Remark',
        hintText: 'Please input the remark for cache',
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

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<ConfigBloc, ConfigState>(builder: (context, state) {
        if (state is ConfigErrorState) {
          return XErrorMessage(
            message: state.message,
            title: 'Get cache config fail',
          );
        }
        final currentConfig = state as ConfigCurrentState;
        var btnText = _isEditting ? 'Save Cache' : 'Add Cache';
        if (currentConfig.isProcessing) {
          btnText = 'Processing...';
        }
        return SingleChildScrollView(
          child: Container(
            margin: EdgeInsets.all(3 * Application.defaultPadding),
            child: Column(
              children: [
                _renderCacheList(currentConfig),
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
                      if ((_formKey.currentState as FormState).validate()) {
                        _addCache(currentConfig);
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
