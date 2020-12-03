///
/// 缓存相关的event
///
abstract class CacheEvent {}

// 缓存删除事件
class CacheRemoveEvent extends CacheEvent {
  final String cache;
  final String key;
  CacheRemoveEvent({
    this.cache,
    this.key,
  });
}
