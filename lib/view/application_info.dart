///
/// 应用信息
///
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/bloc.dart';
import '../config/application.dart';
import '../widget/card.dart';

class _InfoDetail {
  final String name;
  final String value;
  _InfoDetail({
    this.name,
    this.value,
  });
}

@immutable
class ApplicationInfoPage extends StatefulWidget {
  const ApplicationInfoPage({
    Key key,
  }) : super(key: key);
  @override
  _ApplicationInfoPageState createState() => _ApplicationInfoPageState();
}

class _ApplicationInfoPageState extends State<ApplicationInfoPage> {
  final _infoHeight = 25.0;
  @override
  void initState() {
    super.initState();
    BlocProvider.of<ApplicationInfoBloc>(context)
        .add(ApplicationInfoFetchEvent());
  }

  Widget _renderItem(String key, String value) => Row(
        children: [
          Container(
            width: 140,
            height: _infoHeight,
            margin: EdgeInsets.only(
              right: Application.defaultPadding,
            ),
            child: Text(
              key,
              textAlign: TextAlign.right,
              style: TextStyle(
                color: Application.greyBlackColor,
              ),
            ),
          ),
          Container(
            width: 160,
            height: _infoHeight,
            child: Text(value ?? '--'),
          ),
        ],
      );

  Widget _renderBody(ApplicationInfoState state) {
    if (state is ApplicationInfoErrorState) {
      return Center(
        child: Text(state.message),
      );
    }
    final currentState = state as ApplicationInfoCurrentState;
    if (currentState.isProcessing || currentState.info == null) {
      return Center(
        child: Text('Loading...'),
      );
    }

    final info = currentState.info;
    final infos = [
      _InfoDetail(
        name: 'Go Routines',
        value: info.routineCount.toString(),
      ),
      _InfoDetail(
        name: 'Go Max Procs',
        value: info.goMaxProcs.toString(),
      ),
      _InfoDetail(
        name: 'Uptime',
        value: info.uptime,
      ),
      _InfoDetail(
        name: 'Version',
        value: info.version,
      ),
      _InfoDetail(
        name: 'Commit ID',
        value: info.commitID,
      ),
      _InfoDetail(
        name: 'Builded At',
        value: info.buildedAt,
      ),
      _InfoDetail(
        name: 'Go Arch',
        value: info.goarch,
      ),
      _InfoDetail(
        name: 'Go OS',
        value: info.goos,
      ),
      _InfoDetail(
        name: 'Go Version',
        value: info.goVersion,
      ),
    ];
    info.processing?.forEach((key, value) => infos.add(_InfoDetail(
          name: 'Concurrency($key)',
          value: value.toString(),
        )));

    final infosList = <List<_InfoDetail>>[];
    var index = 0;
    final eachLineSize = 3;
    infos.forEach((element) {
      if (index % eachLineSize == 0) {
        infosList.add(<_InfoDetail>[]);
      }
      infosList.last.add(element);
      index++;
    });
    final lastSize = infosList.last.length;
    for (var i = 0; i < eachLineSize - lastSize; i++) {
      infosList.last.add(_InfoDetail(
        name: '',
        value: '',
      ));
    }

    // 		GOARCH       string           `json:"goarch,omitempty"`
    // GOOS         string           `json:"goos,omitempty"`
    // GoVersion    string           `json:"goVersion,omitempty"`
    // Version      string           `json:"version,omitempty"`
    // BuildedAt    string           `json:"buildedAt,omitempty"`
    // CommitID     string           `json:"commitID,omitempty"`
    // Uptime       string           `json:"uptime,omitempty"`
    // GoMaxProcs   int              `json:"goMaxProcs,omitempty"`
    // RoutineCount int              `json:"routineCount,omitempty"`
    // Processing   map[string]int32 `json:"processing,omitempty"`

    return Column(
      children: infosList.map((e) {
        final items = e.map((e) => _renderItem(e.name, e.value)).toList();
        return Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: items,
        );
      }).toList(),
    );
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<ApplicationInfoBloc, ApplicationInfoState>(
        builder: (context, state) => XCard(
          'Application Information',
          _renderBody(state),
        ),
      );
}
