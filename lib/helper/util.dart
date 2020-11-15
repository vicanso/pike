///
/// 常用工具库
///
import 'dart:convert';
import 'package:crypto/crypto.dart';

// 判断是否生产环境
bool isProduct() => bool.fromEnvironment('dart.vm.product');

// 获取base url
String getBaseURL() {
  return 'http://127.0.0.1:9013';
}

// 获取url
String getURL(String url) {
  if (url.startsWith("http")) {
    return url;
  }
  return getBaseURL() + url;
}

// sha256B64 sha256 and conver to base64
String sha256B64(String value) {
  final data = utf8.encode(value);
  return base64.encode(sha256.convert(data).bytes);
}

// hashPassword hash password
String hashPassword(String password) {
  return sha256B64('pike-' + password);
}
