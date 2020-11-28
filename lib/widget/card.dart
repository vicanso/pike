///
/// Card组件
///
import 'package:flutter/material.dart';

import '../config/application.dart';

class XCard extends StatelessWidget {
  final String title;
  final Widget content;
  const XCard(
    this.title,
    this.content, {
    Key key,
  })  : assert(
          title != null,
          'Title can not be null',
        ),
        super(key: key);
  @override
  Widget build(BuildContext context) => Container(
        decoration: BoxDecoration(
          border: Border.all(
            color: Application.defaultBorderColor,
          ),
        ),
        child: Column(
          children: [
            Container(
              width: double.infinity,
              padding: EdgeInsets.all(2 * Application.defaultPadding),
              decoration: BoxDecoration(
                color: Application.greyColor,
                border: Border(
                  bottom: BorderSide(
                    color: Application.defaultBorderColor,
                  ),
                ),
              ),
              child: Text(
                title,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
            Container(
              width: double.infinity,
              padding: EdgeInsets.all(2 * Application.defaultPadding),
              child: content,
            ),
          ],
        ),
      );
}
