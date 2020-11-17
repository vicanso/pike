///
/// 各类自定义的button
///
import 'package:flutter/material.dart';

import '../config/application.dart';

// XFullButton 宽度占满父元素的button
class XFullButton extends StatelessWidget {
  final Widget text;
  final VoidCallback onPressed;
  final EdgeInsetsGeometry padding;
  final EdgeInsetsGeometry margin;
  const XFullButton({
    @required this.text,
    @required this.onPressed,
    this.padding,
    this.margin,
    Key key,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final currentPadding = padding ?? EdgeInsets.all(20.0);
    final currentMargin = margin ?? EdgeInsets.all(15.0);
    return Container(
      width: double.infinity,
      margin: currentMargin,
      child: RaisedButton(
        padding: currentPadding,
        onPressed: onPressed,
        textColor: Application.fontColorOfPrimaryColor,
        color: Theme.of(context).primaryColor,
        child: text,
      ),
    );
  }
}
