///
/// 应用配置
///
import 'package:flutter/material.dart';

import '../router/routes.dart';

class Application {
  static Routes routes;

  // 导航条高度
  static double navigationHeight = 60.0;

  // 主颜色
  static Color primaryColor = Color.fromARGB(255, 3, 0, 28);
  // 主颜色上使用的字体颜色
  static Color fontColorOfPrimaryColor = Colors.white;

  // 次要颜色
  static Color secondaryColor = Colors.grey[100];
  // 次要颜色上使用的字体颜色
  static Color fontColorOfSecondaryColor = Colors.black87;

  // 蓝色
  static Color blueColor = Colors.blueAccent;

  // 主字体颜色
  static Color primaryFontCoolor = Colors.black87;
  // 默认边框颜色
  static Color defaultBorderColor = Colors.grey[200];

  // 常用的几种字体大小
  static double tinyFontSize = 10.0;
  static double smallFontSize = 12.0;
  static double defaultFontSize = 14.0;
  static double normalFontSize = 16.0;
  static double bigFontSize = 18.0;

  // 默认的填充大小
  static double defaultPadding = 10;
}
