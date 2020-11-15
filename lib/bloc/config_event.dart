import 'package:web/model/config.dart';

///
/// 配置事件
///
import '../model/config.dart';

abstract class ConfigEvent {}

// 配置信息拉取
class ConfigFetch extends ConfigEvent {}

// 更新配置信息
class ConfigUpdate extends ConfigEvent {
  final Config config;
  ConfigUpdate({
    this.config,
  });
}
