///
/// 应用信息相关bloc
///
import 'package:bloc/bloc.dart';

import '../config/url.dart' as urls;
import '../helper/request.dart';
import '../helper/util.dart';
import '../model/application_info.dart';
import './application_info_event.dart';
import './application_info_state.dart';

class ApplicationInfoBloc
    extends Bloc<ApplicationInfoEvent, ApplicationInfoState> {
  ApplicationInfoBloc() : super(ApplicationInfoCurrentState());

  @override
  Stream<ApplicationInfoState> mapEventToState(
      ApplicationInfoEvent event) async* {
    // 只有一个event
    yield ApplicationInfoCurrentState(
      processing: true,
    );
    try {
      final resp = await getClient().get(getURL(urls.applicationInfo));
      throwErrorIfFail(resp);
      final info = ApplicationInfo.fromJson(resp.body);
      yield ApplicationInfoCurrentState(
        info: info,
      );
    } on Exception catch (e) {
      yield ApplicationInfoErrorState(
        message: e.toString(),
      );
    }
  }
}
