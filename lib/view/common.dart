///
/// 公共函数
///
import 'package:flutter/material.dart';

import '../config/application.dart';

// createRowItem 生成表单元素
Widget createRowItem(String text) => Padding(
      padding: EdgeInsets.only(
        top: Application.defaultPadding,
        bottom: Application.defaultPadding,
      ),
      child: Text(
        text ?? '--',
        textAlign: TextAlign.center,
      ),
    );

// createRowListItem create row list item
Widget createRowListItem(List<String> arr) => Padding(
      padding: EdgeInsets.only(
        top: Application.defaultPadding,
        bottom: Application.defaultPadding,
      ),
      child: Column(
        children: arr?.map((e) => Text(e))?.toList(),
      ),
    );
