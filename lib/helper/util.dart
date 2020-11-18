///
/// 常用工具库
///
import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:fluttertoast/fluttertoast.dart';

// 判断是否生产环境
const bool isProduct = bool.fromEnvironment('dart.vm.product');

// 获取base url
String getBaseURL() {
  if (isProduct) {
    return '';
  }
  return 'http://127.0.0.1:9013';
}

// 获取url
String getURL(String url) {
  if (url.startsWith('http')) {
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
String hashPassword(String password) => sha256B64('pike-$password');

// showToast show toast
void showToast(String message, {int seconds = 2}) {
  Fluttertoast.showToast(
    msg: message,
    toastLength: Toast.LENGTH_SHORT,
    gravity: ToastGravity.CENTER,
    timeInSecForIosWeb: seconds,
    webBgColor: '#edf2fc',
  );
}

// showErrorMessage show error message toast
void showErrorMessage(String error) {
  Fluttertoast.showToast(
    msg: error,
    toastLength: Toast.LENGTH_SHORT,
    gravity: ToastGravity.CENTER,
    timeInSecForIosWeb: 2,
    webBgColor: '#fe6c6f',
  );
}

// createNumberValidator create number validator
String Function(String) createNumberValidator(String message) {
  final reg = RegExp(r'^\d+$');

  final fn = (String value) {
    if (value == null || !reg.hasMatch(value) || int.parse(value) == 0) {
      return message;
    }
    return null;
  };
  return fn;
}
