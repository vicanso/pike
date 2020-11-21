///
/// 配置信息相关的bloc
///
import 'package:bloc/bloc.dart';

import '../config/url.dart' as urls;
import '../helper/request.dart';
import '../helper/util.dart';
import '../model/config.dart';
import './config_event.dart';
import './config_state.dart';

class ConfigBloc extends Bloc<ConfigEvent, ConfigState> {
  ConfigBloc() : super(ConfigCurrentState());

  @override
  Stream<ConfigState> mapEventToState(ConfigEvent event) async* {
    if (event is ConfigFetch) {
      yield ConfigCurrentState(
        processing: true,
      );
      try {
        final resp = await getClient().get(getURL(urls.config));
        throwErrorIfFail(resp);
        final c = Config.fromJson(resp.body);
        yield ConfigCurrentState(
          config: c,
        );
      } on Exception catch (e) {
        yield ConfigErrorState(
          message: e.toString(),
        );
      }
      return;
    }

    if (event is ConfigUpdate) {
      if (state is ConfigCurrentState) {
        yield (state as ConfigCurrentState).copyWith(
          processing: true,
        );
      }
      try {
        final data = event.config.toJson();

        final resp = await getClient().put(
          '${getURL(urls.config)}?delay=${event.delay ?? ""}',
          body: data,
          headers: {
            'Content-Type': 'application/json',
          },
        );
        throwErrorIfFail(resp);
        final c = Config.fromJson(resp.body);
        yield ConfigCurrentState(
          config: c,
        );
      } on Exception catch (e) {
        yield ConfigErrorState(
          message: e.toString(),
        );
      }
      return;
    }
  }
}
