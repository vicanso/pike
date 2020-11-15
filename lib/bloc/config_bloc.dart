///
/// 配置信息相关的bloc
///
import 'package:bloc/bloc.dart';
import 'package:web/helper/util.dart';

import './config_event.dart';
import './config_state.dart';
import '../config/url.dart' as urls;
import '../helper/request.dart';
import '../model/config.dart';

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
      } catch (e) {
        yield ConfigErrorState(
          message: e.toString(),
        );
      }
      return;
    }
  }
}
