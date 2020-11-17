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

@immutable
class CompressPage extends StatefulWidget {
  const CompressPage({
    Key key,
  }) : super(key: key);
  @override
  _CompressPageState createState() => _CompressPageState();
}

class _CompressPageState extends State<CompressPage> {
  final GlobalKey _formKey = GlobalKey<FormState>();
  final TextEditingController _nameController = TextEditingController();
  final TextEditingController _gzipController = TextEditingController();
  final TextEditingController _brController = TextEditingController();
  final TextEditingController _remarkController = TextEditingController();
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

  String _getLevelDesc(Map<String, int> levels, String name) {
    final value = levels[name];
    if (value == null) {
      return '--';
    }
    return value.toString();
  }

  String _getLevel(Map<String, int> levels, String name) {
    final value = levels[name];
    if (value == null) {
      return null;
    }
    return value.toString();
  }

  void _reset() {
    _nameController.clear();
    _gzipController.clear();
    _brController.clear();
    _remarkController.clear();
  }

  Widget _createRowItem(String text) => Padding(
        padding: EdgeInsets.only(
          top: Application.defaultPadding,
          bottom: Application.defaultPadding,
          left: Application.defaultPadding,
        ),
        child: Text(
          text ?? '--',
          textAlign: TextAlign.center,
        ),
      );

  void _deleteCompress(ConfigCurrentState state, String name) {
    final compressList = <CompressConfig>[];

    if (!state.config.validateForDelete('compress', name)) {
      showErrorMessage('$name is used, it can not be deleted');
      return;
    }

    state.config.compresses?.forEach((element) {
      if (element.name != name) {
        compressList.add(element);
      }
    });
    _configBloc.add(ConfigUpdate(
      config: state.config.copyWith(
        compresses: compressList,
      ),
    ));
  }

  Widget _renderCurrentCompress(ConfigCurrentState state) {
    final gzipName = 'Gzip';
    final brName = 'Br';
    final rows = <TableRow>[
      TableRow(
        children: [
          _createRowItem('Name'),
          _createRowItem(gzipName),
          _createRowItem(brName),
          _createRowItem('Remark'),
          _createRowItem('Operations'),
        ],
      ),
    ];
    state.config.compresses?.forEach((element) {
      rows.add(TableRow(
        children: [
          _createRowItem(element.name),
          _createRowItem(_getLevelDesc(element.levels, gzipName.toLowerCase())),
          _createRowItem(_getLevelDesc(element.levels, brName.toLowerCase())),
          _createRowItem(element.remark),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              TextButton(
                onPressed: () {
                  _reset();
                  _nameController.value = TextEditingValue(text: element.name);
                  _gzipController.value = TextEditingValue(
                      text: _getLevel(
                    element.levels,
                    gzipName.toLowerCase(),
                  ));
                  _brController.value = TextEditingValue(
                      text: _getLevel(
                    element.levels,
                    brName.toLowerCase(),
                  ));
                  _remarkController.value = TextEditingValue(
                    text: element.remark,
                  );
                  setState(() {
                    _mode = _updateMode;
                  });
                },
                child: Text('Update'),
              ),
              TextButton(
                onPressed: () {
                  _deleteCompress(state, element.name);
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
        4: FixedColumnWidth(150),
      },
      border: TableBorder.all(
        color: Application.primaryColor.withAlpha(60),
      ),
      children: rows,
    );
  }

  String _validateNumber(String value) {
    final reg = RegExp(r'\d+');
    if (value == null || !reg.hasMatch(value) || int.parse(value) == 0) {
      return 'compress level should be number and gt 0';
    }
    return null;
  }

  Widget _renderEditor() {
    if (!_isEditting) {
      return Container();
    }
    final formItems = <Widget>[];
    formItems.add(TextFormField(
      autofocus: true,
      readOnly: _isUpdateding,
      controller: _nameController,
      decoration: InputDecoration(
        labelText: 'Name',
        hintText: 'Please input the name of compress',
      ),
      validator: (v) => v.trim().isNotEmpty ? null : 'name can not be null',
    ));

    formItems.add(TextFormField(
      controller: _gzipController,
      decoration: InputDecoration(
        labelText: 'Gzip Level',
        hintText: 'Please input the compress level of gzip(1-9)',
      ),
      validator: _validateNumber,
    ));

    formItems.add(TextFormField(
      controller: _brController,
      decoration: InputDecoration(
        labelText: 'Br Level',
        hintText: 'Please input the compress level of br(1-11)',
      ),
      validator: _validateNumber,
    ));

    formItems.add(TextFormField(
      controller: _remarkController,
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

  void _addCompress(ConfigCurrentState state) {
    final name = _nameController.text;
    final levels = <String, int>{
      'gzip': int.parse(_gzipController.text),
      'br': int.parse(_brController.text),
    };
    final compressConfig = CompressConfig(
      name: name,
      levels: levels,
      remark: _remarkController.text?.trim(),
    );
    // var found = false;
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
        return SingleChildScrollView(
          child: Container(
            margin: EdgeInsets.all(3 * Application.defaultPadding),
            child: Column(
              children: [
                _renderCurrentCompress(currentConfig),
                _renderEditor(),
                XFullButton(
                  margin: EdgeInsets.only(
                    top: 3 * Application.defaultPadding,
                  ),
                  onPressed: () {
                    if (currentConfig.isProcessing) {
                      return;
                    }
                    if (_isEditting) {
                      if ((_formKey.currentState as FormState).validate()) {
                        _addCompress(currentConfig);
                      }
                      return;
                    }
                    _reset();
                    setState(() {
                      _mode = _editMode;
                    });
                  },
                  text: Text(_isEditting ? 'Save Compress' : 'Add Compress'),
                )
              ],
            ),
          ),
        );
      });
}
