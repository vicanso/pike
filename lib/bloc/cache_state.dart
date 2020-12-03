///
/// 缓存相关的state
///
import 'package:equatable/equatable.dart';

abstract class CacheState extends Equatable {}

// 缓存列表信息
class CacheListState extends CacheState {
  final bool processing;
  CacheListState({
    this.processing,
  });

  bool get isProcessing => processing != null && processing;

  @override
  List<Object> get props => [processing];

  @override
  String toString() => 'CacheListState(processing: $processing)';

  CacheListState copyWith({
    bool processing,
  }) =>
      CacheListState(
        processing: processing ?? this.processing,
      );
}

// 缓存列表出错信息
class CacheErrorState extends CacheState {
  final String message;
  CacheErrorState({
    this.message,
  });

  @override
  List<Object> get props => [message];

  @override
  String toString() => 'CacheErrorState(message: $message)';
}
