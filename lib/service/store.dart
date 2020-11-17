///
/// 本地缓存key/value
///
import 'package:shared_preferences/shared_preferences.dart';

SharedPreferences _prefs;
// _getInstance get instance
Future<void> init() async => _prefs = await SharedPreferences.getInstance();

class Store {
  final String prefix;
  // Store
  Store({
    this.prefix,
  });

  String _getKey(String key) => '$prefix$key';

  // getString get string from store
  String getString(String key) => _prefs.getString(_getKey(key)) ?? '';

  // setString set string to store
  Future<bool> setString(String key, String value) =>
      _prefs.setString(_getKey(key), value);

  // getStringList get string list from store
  List<String> getStringList(String key) => _prefs.getStringList(key);

  // setStringList set string list to store
  Future<bool> setStringList(String key, List<String> value) =>
      _prefs.setStringList(key, value);
}
