///
/// 用户信息相关bloc
///
import 'package:bloc/bloc.dart';

import './user_event.dart';
import './user_state.dart';
import '../config/url.dart' as urls;
import '../helper/request.dart';
import '../helper/util.dart';
import '../model/user.dart';

class UserBloc extends Bloc<UserEvent, UserState> {
  UserBloc() : super(UserMeState());

  @override
  Stream<UserState> mapEventToState(UserEvent event) async* {
    if (event is UserMeFetch) {
      // 切换状态为处理中
      yield UserMeState(
        processing: true,
      );
      try {
        final resp = await getClient().get(getURL(urls.userMe));
        throwErrorIfFail(resp);
        final user = User.fromJson(resp.body);
        yield UserMeState(
          user: user,
        );
      } catch (e) {
        yield UserErrorState(
          message: e.toString(),
        );
      }
      return;
    }

    if (event is UserLogin) {
      yield UserMeState(
        processing: true,
      );
      try {
        // 登录时对密码进去sha256处理
        final resp = await getClient().post(
          getURL(urls.userLogin),
          body: {
            "account": event.account,
            "password": hashPassword(event.password),
          },
        );
        throwErrorIfFail(resp);
        final user = User.fromJson(resp.body);
        yield UserMeState(
          user: user,
        );
      } catch (e) {
        yield UserErrorState(
          message: e.toString(),
        );
      }
      return;
    }
  }
}
