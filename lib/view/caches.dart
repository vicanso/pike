///
/// 缓存列表页
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../helper/util.dart';
import '../widget/error_message.dart';

@immutable
class CachesPage extends StatefulWidget {
  const CachesPage({
    Key key,
  }) : super(key: key);
  @override
  _CachesPageState createState() => _CachesPageState();
}

class _CachesPageState extends State<CachesPage> {
  final _cacheURLCtrl = TextEditingController();

  CacheBloc _cacheBloc;
  @override
  void initState() {
    super.initState();
    _cacheBloc = BlocProvider.of<CacheBloc>(context);
  }

  String _method = 'GET';
  Widget _renderDeleteCache(CacheState state) {
    final divider = Container(
      height: Application.defaultPadding,
      width: 2 * Application.defaultPadding,
    );
    var isProcessing = false;
    if (state is CacheListState) {
      isProcessing = state.isProcessing;
    }

    return Container(
      margin: EdgeInsets.only(
        top: 3 * Application.defaultPadding,
      ),
      child: Row(
        children: [
          DropdownButton<String>(
            value: _method,
            itemHeight: 64,
            // icon: Icon(Icons.arrow_downward),
            onChanged: (newValue) {
              setState(() {
                _method = newValue;
              });
            },
            items: [
              'GET',
              'HEAD',
            ]
                .map<DropdownMenuItem<String>>(
                    (String value) => DropdownMenuItem<String>(
                          value: value,
                          child: Text(value),
                        ))
                .toList(),
          ),
          divider,
          Expanded(
            child: TextFormField(
              controller: _cacheURLCtrl,
              decoration: InputDecoration(
                labelText: 'Cache URL',
                hintText:
                    'Please input the cache url, e.g.: http://test.com/user?vip=1',
              ),
            ),
          ),
          divider,
          RaisedButton(
            padding: EdgeInsets.all(20.0),
            textColor: Application.fontColorOfPrimaryColor,
            color: Theme.of(context).primaryColor,
            child: Text(isProcessing ? 'Deleting' : 'Delete'),
            onPressed: () {
              final key = _cacheURLCtrl.text?.trim();
              if (key == null || key.isEmpty) {
                showErrorMessage('Cache url can not be empty');
                return;
              }
              final info = Uri.parse(key);
              var url = info.path;
              if (info.query != null && info.query.isNotEmpty) {
                url += '?${info.query}';
              }
              var hostPort = info.host;
              if (info.hasPort) {
                hostPort += ':${info.port.toString()}';
              }
              final cacheKey = '$_method $hostPort $url';
              _cacheBloc.add(CacheRemoveEvent(
                key: cacheKey,
              ));
            },
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<CacheBloc, CacheState>(builder: (context, state) {
        if (state is CacheErrorState) {
          return XErrorMessage(
            message: state.message,
            title: 'Get cache fail',
          );
        }
        return Container(
          margin: EdgeInsets.all(3 * Application.defaultPadding),
          child: _renderDeleteCache(state),
        );
      });
}
