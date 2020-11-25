///
/// 自定义选择器
///
import 'package:flutter/material.dart';

import '../config/application.dart';

// 表单中的选择器
class XFormSelector extends StatelessWidget {
  final String title;
  final String value;
  final List<String> values;
  final List<String> options;
  final ValueChanged<String> onChanged;
  final bool toggled;
  final bool mutiple;

  const XFormSelector({
    @required this.title,
    @required this.options,
    @required this.onChanged,
    this.value,
    this.toggled,
    this.mutiple,
    this.values,
    Key key,
  }) : super(key: key);

  bool get _isSupportToggled => toggled != null && toggled;
  bool get _isSupportMultiple => mutiple != null && mutiple;
  @override
  Widget build(BuildContext context) {
    final items = <Widget>[
      Text(title),
      Container(
        width: Application.defaultPadding,
      ),
    ];
    options.forEach((element) {
      Color iconColor = Colors.blue;
      Color textColor = Colors.blue;
      if (_isSupportMultiple) {
        if (!(values?.contains(element) ?? false)) {
          iconColor = Colors.grey;
          textColor = Application.fontColorOfSecondaryColor;
        }
      } else {
        if (value != element) {
          iconColor = Colors.grey;
          textColor = Application.fontColorOfSecondaryColor;
        }
      }
      items.add(RaisedButton(
        color: Colors.white,
        textColor: iconColor,
        onPressed: () {
          if (_isSupportMultiple) {
            final result = <String>[];
            result.addAll(values ?? <String>[]);
            if (result.contains(element)) {
              if (_isSupportToggled) {
                result.remove(element);
              }
            } else {
              result.add(element);
            }
            onChanged(result.join(','));
            return;
          }
          if (element == value) {
            // 如果支持切换选中
            if (_isSupportToggled) {
              onChanged('');
            }
            return;
          }
          onChanged(element);
        },
        child: Row(
          children: [
            Icon(
              Icons.check,
              size: Application.smallFontSize,
              color: iconColor,
            ),
            Container(
              width: 3,
            ),
            Text(
              element,
              style: TextStyle(
                color: textColor,
              ),
            ),
          ],
        ),
      ));
      items.add(Container(
        width: Application.defaultPadding,
      ));
    });
    return Row(
      children: items,
    );
  }
}
