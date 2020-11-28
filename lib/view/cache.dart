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
import '../widget/table.dart';

@immutable
class CachePage extends StatefulWidget {
  const CachePage({
    Key key,
  }) : super(key: key);
  @override
  _CachePageState createState() => _CachePageState();
}

class _CachePageState extends State<CachePage> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtrl = TextEditingController();
  final _sizeCtrl = TextEditingController();
  final _hitForPassCtrl = TextEditingController();
  final _remarkCtrl = TextEditingController();

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

  bool get _isUpdating => _mode == _updateMode;

  // _reset 重置表单所有元素
  void _reset() {
    _nameCtrl.clear();
    _sizeCtrl.clear();
    _hitForPassCtrl.clear();
    _remarkCtrl.clear();
  }

  // _fillEditor 填充编辑数据
  void _fillEditor(CacheConfig element) {
    _nameCtrl.value = TextEditingValue(text: element.name);
    _sizeCtrl.value = TextEditingValue(text: element.size.toString());
    _hitForPassCtrl.value = TextEditingValue(text: element.hitForPass);
    _remarkCtrl.value = TextEditingValue(text: element.remark ?? '');
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

  // _renderCacheList 渲染当前缓存服务列表
  Widget _renderCacheList(ConfigCurrentState state) {
    // 表格内容
    final contents = state.config.caches
        ?.map((e) => [
              e.name,
              numberFormat.format(e.size),
              e.hitForPass,
              e.remark,
            ])
        ?.toList();
    final doUpdate = (int index) {
      final element = state.config.caches.elementAt(index);
      // 重置当前数据，并将需要更新的配置填充
      _reset();
      _fillEditor(element);

      setState(() {
        _mode = _updateMode;
      });
    };
    final doDelete = (int index) {
      final element = state.config.caches.elementAt(index);
      _deleteCache(state, element.name);
    };
    return XConfigTable(
      headers: [
        'Name',
        'Size',
        'Hit For Pass',
        'Remark',
      ],
      contents: contents,
      onUpdate: doUpdate,
      onDelete: doDelete,
      columnWidths: <String, double>{
        'Size': 100,
        'Hit For Pass': 130,
      },
    );
  }

  // _addCache 添加缓存服务，如果存在相同服务，则替换
  void _addCache(ConfigCurrentState state) {
    final cacheConfig = CacheConfig(
      name: _nameCtrl.text?.trim(),
      size: int.parse(_sizeCtrl.text),
      hitForPass: _hitForPassCtrl.text?.trim(),
      remark: _remarkCtrl.text?.trim(),
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
      readOnly: _isUpdating,
      controller: _nameCtrl,
      decoration: InputDecoration(
        labelText: 'Name',
        hintText: 'Please input the name of cache',
      ),
      validator: (v) => v.trim().isNotEmpty ? null : 'name can not be null',
    ));

    // 缓存大小
    formItems.add(TextFormField(
      controller: _sizeCtrl,
      decoration: InputDecoration(
        labelText: 'Size',
        hintText: 'Please input the size of cache, e.g.: 51200',
      ),
      validator:
          createNumberValidator('size of cache should be number and gt 0'),
    ));

    // hit for pass时长
    formItems.add(TextFormField(
      controller: _hitForPassCtrl,
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
      controller: _remarkCtrl,
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
                      if (_formKey.currentState.validate()) {
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
