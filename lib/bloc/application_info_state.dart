///
/// 应用信息相关的state
///
import 'package:equatable/equatable.dart';

import '../model/application_info.dart';

abstract class ApplicationInfoState extends Equatable {}

// ApplicationInfoCurrentState 应用程序信息
class ApplicationInfoCurrentState extends ApplicationInfoState {
  final ApplicationInfo info;
  final bool processing;
  ApplicationInfoCurrentState({
    this.info,
    this.processing,
  });

  // isProcessing 是否正在拉取用户信息
  bool get isProcessing => processing != null && processing;

  @override
  List<Object> get props => [info, processing];

  @override
  String toString() =>
      'ApplicationInfoCurrentState(info: $info, processing: $processing)';
}

class ApplicationInfoErrorState extends ApplicationInfoState {
  final String message;
  ApplicationInfoErrorState({
    this.message,
  });

  @override
  List<Object> get props => [message];

  @override
  String toString() => 'ApplicationInfoErrorState(message: $message)';
}
