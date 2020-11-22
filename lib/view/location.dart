///
/// location配置页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../helper/util.dart';
import '../model/config.dart';
import '../widget/button.dart';
import '../widget/error_message.dart';
import '../widget/selector.dart';
import './common.dart';

@immutable
class LocationPage extends StatefulWidget {
  const LocationPage({
    Key key,
  }) : super(key: key);
  @override
  _LocationPageState createState() => _LocationPageState();
}

class _LocationPageState extends State<LocationPage> {
  final _formKey = GlobalKey<FormState>();

  final _nameCtrl = TextEditingController();
  // url前置配置
  final _prefixCtrlList = <TextEditingController>[
    TextEditingController(),
  ];
  // host配置
  final _hostCtrlList = <TextEditingController>[
    TextEditingController(),
  ];
  // url重写配置
  final _rewriteCtrlList = <TextEditingController>[
    TextEditingController(),
  ];
  // 响应头配置
  final _respHeaderCtrlList = <TextEditingController>[
    TextEditingController(),
  ];
  // 请求头配置
  final _reqHeaderCtrlList = <TextEditingController>[
    TextEditingController(),
  ];

  // 超时配置
  final _proxyTimeoutCtrl = TextEditingController();
  final _remarkCtrl = TextEditingController();

  String _upstream = '';

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
  void _reset() {
    _nameCtrl.clear();
    _upstream = '';

    [
      _prefixCtrlList,
      _hostCtrlList,
      _rewriteCtrlList,
      _respHeaderCtrlList,
      _reqHeaderCtrlList,
    ].forEach((element) {
      element.clear();
      element.add(TextEditingController());
    });

    _proxyTimeoutCtrl.clear();
    _remarkCtrl.clear();
  }

  // _fillList 填充列表
  void _fillList(List<TextEditingController> ctrls, List<String> values) {
    if (values == null || values.isEmpty) {
      return;
    }
    ctrls.clear();
    values.forEach((element) {
      final ctrl = TextEditingController(
        text: element ?? '',
      );
      ctrls.add(ctrl);
    });
  }

  // _fillEditor 填充数据
  void _fillEditor(LocationConfig element) {
    _nameCtrl.value = TextEditingValue(text: element.name ?? '');
    _upstream = element.upstream;
    _fillList(_prefixCtrlList, element.prefixes);
    _fillList(_hostCtrlList, element.hosts);
    _fillList(_rewriteCtrlList, element.rewrites);
    _fillList(_respHeaderCtrlList, element.respHeaders);
    _fillList(_reqHeaderCtrlList, element.reqHeaders);
    _proxyTimeoutCtrl.value =
        TextEditingValue(text: element.proxyTimeout ?? '');
    _remarkCtrl.value = TextEditingValue(text: element.remark ?? '');
  }

  // _renderLocationEditor 渲染location编辑表单
  Widget _renderLocationEditor(ConfigCurrentState state) {
    if (!_isEditting) {
      return Container();
    }
    final formItems = <Widget>[];
    // 名称
    formItems.add(TextFormField(
      autofocus: true,
      readOnly: _isUpdateding,
      controller: _nameCtrl,
      decoration: InputDecoration(
        labelText: 'Name',
        hintText: 'Please input the name of location',
      ),
      validator: (v) => v.trim().isNotEmpty ? null : 'Name can not be null',
    ));

    final upstreamList = state.config.upstreams?.map((e) => e.name)?.toList();

    // upstream选择器（如果无则不需要展示）
    formItems.add(Container(
      margin: EdgeInsets.only(
        top: Application.defaultPadding,
      ),
      child: XFormSelector(
        title: 'Upstream',
        value: _upstream ?? '',
        options: upstreamList,
        onChanged: (String upstream) {
          setState(() {
            _upstream = upstream;
          });
        },
      ),
    ));

    // url prefix编辑列表
    _prefixCtrlList.forEach((element) {
      formItems.add(TextFormField(
        controller: element,
        decoration: InputDecoration(
          labelText: 'Prefix',
          hintText: 'Please input the url prefix, e.g.: /api',
        ),
        validator: (v) {
          if (v == null || v.isEmpty || v.startsWith('/')) {
            return null;
          }
          return 'Prefix is invalid';
        },
      ));
    });
    // 添加更多的prefix
    formItems.add(Container(
      child: XFullButton(
        text: Text('Add More Prefix'),
        padding: EdgeInsets.all(1.5 * Application.defaultPadding),
        margin: EdgeInsets.only(
          top: Application.defaultPadding,
        ),
        onPressed: () {
          setState(() {
            _prefixCtrlList.add(TextEditingController());
          });
        },
      ),
    ));

    // host编辑列表
    _hostCtrlList.forEach((element) {
      formItems.add(TextFormField(
        controller: element,
        decoration: InputDecoration(
          labelText: 'Host',
          hintText: 'Please input the host of request, e.g.: test.com',
        ),
      ));
    });
    // 添加更多的host
    formItems.add(Container(
      child: XFullButton(
        text: Text('Add More Host'),
        padding: EdgeInsets.all(1.5 * Application.defaultPadding),
        margin: EdgeInsets.only(
          top: Application.defaultPadding,
        ),
        onPressed: () {
          setState(() {
            _hostCtrlList.add(TextEditingController());
          });
        },
      ),
    ));

    // url rewrite编辑列表
    _rewriteCtrlList.forEach((element) {
      formItems.add(TextFormField(
        controller: element,
        decoration: InputDecoration(
          labelText: 'Rewrite',
          hintText: 'Please input the url rewrite, e.g.: /api/*:/\$1',
        ),
        validator: (v) {
          if (v == null || v.isEmpty) {
            return null;
          }
          if (v.contains(':')) {
            return null;
          }
          return 'Rewrite is invalid';
        },
      ));
    });
    // 添加更多rewrite
    formItems.add(Container(
      child: XFullButton(
        text: Text('Add More Rewrite'),
        padding: EdgeInsets.all(1.5 * Application.defaultPadding),
        margin: EdgeInsets.only(
          top: Application.defaultPadding,
        ),
        onPressed: () {
          setState(() {
            _rewriteCtrlList.add(TextEditingController());
          });
        },
      ),
    ));

    // 响应头
    _respHeaderCtrlList.forEach((element) {
      formItems.add(TextFormField(
        controller: element,
        decoration: InputDecoration(
            labelText: 'Resp Header',
            hintText: 'Please input the response header, e.g.: X-Resp-Id:1'),
      ));
    });
    // 添加更多的响应头
    formItems.add(Container(
      child: XFullButton(
        text: Text('Add More Response Header'),
        padding: EdgeInsets.all(1.5 * Application.defaultPadding),
        margin: EdgeInsets.only(
          top: Application.defaultPadding,
        ),
        onPressed: () {
          setState(() {
            _respHeaderCtrlList.add(TextEditingController());
          });
        },
      ),
    ));

    // 请求头
    _reqHeaderCtrlList.forEach((element) {
      formItems.add(TextFormField(
        controller: element,
        decoration: InputDecoration(
          labelText: 'Req Header',
          hintText: 'Please int the request header, e.g.: X-Req-Id:1',
        ),
      ));
    });
    // 添加更多的请求头
    formItems.add(Container(
      child: XFullButton(
        text: Text('Add More Request Header'),
        padding: EdgeInsets.all(1.5 * Application.defaultPadding),
        margin: EdgeInsets.only(
          top: Application.defaultPadding,
        ),
        onPressed: () {
          setState(() {
            _reqHeaderCtrlList.add(TextEditingController());
          });
        },
      ),
    ));

    // 超时配置
    formItems.add(TextFormField(
      controller: _proxyTimeoutCtrl,
      decoration: InputDecoration(
        labelText: 'Proxy Timeout',
        hintText: 'Please input the timeout of proxy request(10s, 1m)',
      ),
      validator: (value) {
        final reg = RegExp(r'^\d+[sm]$');
        if (value == null || !reg.hasMatch(value)) {
          return 'timeout is invalid';
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
        hintText: 'Please input the remark for location',
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

  // _renderLocationList 渲染location 列表
  Widget _renderLocationList(ConfigCurrentState state) {
    // 表头
    final rows = <TableRow>[
      TableRow(
        children: [
          createRowItem('Name'),
          createRowItem('Upstream'),
          createRowItem('Prefixes'),
          createRowItem('Hosts'),
          createRowItem('Rewrites'),
          createRowItem('Resp Headers'),
          createRowItem('Req Headers'),
          createRowItem('Proxy Timeout'),
          createRowItem('Remark'),
          createRowItem('Operations'),
        ],
      ),
    ];
    state.config.locations?.forEach((element) {
      rows.add(TableRow(
        children: [
          createRowItem(element.name),
          createRowItem(element.upstream),
          createRowListItem(element.prefixes),
          createRowListItem(element.hosts),
          createRowListItem(element.rewrites),
          createRowListItem(element.respHeaders),
          createRowListItem(element.reqHeaders),
          createRowItem(element.proxyTimeout),
          createRowItem(element.remark),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              TextButton(
                onPressed: () {
                  // 重置当前数据，并将需要更新的配置填充
                  _reset();
                  _fillEditor(element);

                  setState(() {
                    _mode = _updateMode;
                  });
                },
                child: Text('Update'),
              ),
              TextButton(
                onPressed: () {
                  _deleteLocation(state, element.name);
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
        1: FixedColumnWidth(120),
        2: FixedColumnWidth(120),
        7: FixedColumnWidth(100),
        9: FixedColumnWidth(150),
      },
      border: TableBorder.all(
        color: Application.primaryColor.withAlpha(60),
      ),
      children: rows,
    );
  }

  List<String> _getValues(List<TextEditingController> list) {
    final values = <String>[];
    list.forEach((element) {
      final value = element.text?.trim();
      if (value != null && value.isNotEmpty) {
        values.add(value);
      }
    });
    return values;
  }

  // _deleteLocation 删除location
  void _deleteLocation(ConfigCurrentState state, String name) {
    // 校验是否有已使用该location
    if (!state.config.validateForDelete('location', name)) {
      showErrorMessage('$name is used, it can not be deleted');
      return;
    }
    final locationList = <LocationConfig>[];
    state.config.locations?.forEach((element) {
      if (element.name != name) {
        locationList.add(element);
      }
    });
    // 更新配置
    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        locations: locationList,
      ),
    ));
  }

  // _addLocation 添加location配置，如果添加的配置已存在，则替换
  void _addLocation(ConfigCurrentState state) {
    if (_upstream == null || _upstream.isEmpty) {
      showErrorMessage('upstream is required');
      return;
    }
    final name = _nameCtrl.text?.trim();
    final locationConfig = LocationConfig(
      name: name,
      upstream: _upstream,
      prefixes: _getValues(_prefixCtrlList),
      hosts: _getValues(_hostCtrlList),
      rewrites: _getValues(_rewriteCtrlList),
      respHeaders: _getValues(_respHeaderCtrlList),
      reqHeaders: _getValues(_reqHeaderCtrlList),
      proxyTimeout: _proxyTimeoutCtrl.text?.trim(),
      remark: _remarkCtrl.text?.trim(),
    );

    final locationList = <LocationConfig>[];
    state.config.locations?.forEach((element) {
      if (element.name != name) {
        locationList.add(element);
      }
    });
    locationList.add(locationConfig);
    _configBloc.add(ConfigUpdate(
        config: state.config.copyWith(
      locations: locationList,
    )));
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
            title: 'Get location config fail',
          );
        }
        final currentConfig = state as ConfigCurrentState;
        var btnText = _isEditting ? 'Save Upstream' : 'Add Upstream';
        if (currentConfig.isProcessing) {
          btnText = 'Processing...';
        }
        return SingleChildScrollView(
          child: Container(
            margin: EdgeInsets.all(3 * Application.defaultPadding),
            child: Column(children: [
              _renderLocationList(currentConfig),
              _renderLocationEditor(currentConfig),
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
                      _addLocation(currentConfig);
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
            ]),
          ),
        );
      });
}
