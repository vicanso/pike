///
/// 缓存信息相关的bloc
///
import 'package:bloc/bloc.dart';

import '../config/url.dart' as urls;
import '../helper/request.dart';
import '../helper/util.dart';
import './cache_event.dart';
import './cache_state.dart';

class CacheBloc extends Bloc<CacheEvent, CacheState> {
  CacheBloc() : super(CacheListState());

  @override
  Stream<CacheState> mapEventToState(CacheEvent event) async* {
    if (event is CacheRemoveEvent) {
      final key = event.key;
      yield CacheListState(
        processing: true,
      );

      try {
        final url = '${getURL(urls.cache)}?key=${Uri.encodeComponent(key)}';
        final resp = await getClient().delete(url);
        throwErrorIfFail(resp);
        yield CacheListState(
          processing: false,
        );
      } on Exception catch (e) {
        yield CacheErrorState(
          message: e.toString(),
        );
      }
      return;
    }
  }
}
