///
/// 配置相关的state
///
import 'package:equatable/equatable.dart';

import '../model/config.dart';

abstract class ConfigState extends Equatable {}

// 配置信息
class ConfigCurrentState extends ConfigState {
  final Config config;
  final bool processing;
  ConfigCurrentState({
    this.config,
    this.processing,
  });
  // isProcessing 是否正在拉取配置信息
  bool get isProcessing => processing != null && processing;

  @override
  List<Object> get props => [config, processing];

  @override
  String toString() =>
      'ConfigCurrentState(config: $config, processing: $processing)';
}

// 拉取配置信息出错
class ConfigErrorState extends ConfigState {
  final String message;
  ConfigErrorState({
    this.message,
  });

  @override
  List<Object> get props => [message];

  @override
  String toString() => 'ConfigErrorState(message: $message)';
}
