///
/// 本地缓存key/value
///
import 'package:shared_preferences/shared_preferences.dart';

SharedPreferences _prefs;
// _getInstance get instance
Future<void> init() async => _prefs = await SharedPreferences.getInstance();
