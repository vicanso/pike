///
/// HTTP请求库
///
import 'dart:convert' as convert;
import 'package:http/browser_client.dart';
import 'package:http/http.dart' as http;

class PikeClient extends http.BaseClient {
  final http.Client client;
  PikeClient({
    this.client,
  });

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) =>
      client.send(request);
}

http.Client initClient() {
  final c = http.Client();
  if (c is BrowserClient) {
    c.withCredentials = true;
  }
  return c;
}

final _c = PikeClient(
  client: initClient(),
);

PikeClient getClient() => _c;

void throwErrorIfFail(http.Response resp) {
  if (resp.statusCode >= 400) {
    final m = convert.jsonDecode(resp.body);
    String message;
    if (m['message'] != null && m['message'] is String) {
      message = m['message'] as String;
    }
    if (message == null || message.isEmpty) {
      message = 'Unknown Error';
    }
    throw Exception(message);
  }
}
