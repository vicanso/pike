///
/// 自定义的出错提示
///
import 'package:flutter/material.dart';

import '../config/application.dart';

class XErrorMessage extends StatelessWidget {
  final String message;
  final String title;
  const XErrorMessage({
    @required this.message,
    @required this.title,
    Key key,
  }) : super(key: key);
  @override
  Widget build(BuildContext context) {
    // ErrorOutline
    final iconWidget = Container(
      margin: EdgeInsets.only(
        top: 40,
        bottom: 20,
      ),
      child: Icon(
        Icons.error_outline,
        size: 80,
        color: Colors.red[900],
      ),
    );
    final titleWidget = Container(
      width: double.infinity,
      margin: EdgeInsets.all(Application.defaultPadding),
      child: Text(
        title,
        textAlign: TextAlign.center,
        style: TextStyle(
          fontSize: Application.bigFontSize,
        ),
      ),
    );
    final contentWidget = Container(
      margin: EdgeInsets.all(Application.defaultPadding),
      child: Text(message),
    );

    return Column(
      children: <Widget>[
        iconWidget,
        titleWidget,
        contentWidget,
      ],
    );
  }
}

class XErrorTips extends StatelessWidget {
  final String message;
  const XErrorTips({
    @required this.message,
    Key key,
  }) : super(key: key);
  @override
  Widget build(BuildContext context) => Row(
        children: [
          Icon(
            Icons.error_outline,
            color: Colors.red[900],
          ),
          Container(
            width: 10,
          ),
          Expanded(
            child: Text(
              message,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      );
}
