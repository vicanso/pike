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
  final ValueChanged<String> onChanged;

  const XFormSelector({
    @required this.title,
    @required this.value,
    @required this.values,
    @required this.onChanged,
    Key key,
  }) : super(key: key);
  @override
  Widget build(BuildContext context) {
    final items = values
        .map<DropdownMenuItem<String>>((String v) => DropdownMenuItem<String>(
              value: v,
              child: Text(v),
            ))
        .toList();
    return Row(
      children: [
        Text(title),
        Container(
          width: Application.defaultPadding,
        ),
        DropdownButton(
          value: value,
          items: items,
          onChanged: onChanged,
        )
      ],
    );
  }
}
