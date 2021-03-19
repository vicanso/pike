# pike admin

最新版本flutter2已默认支持web，使用最新版本则可。

默认stable分支不支持web，因此需要修改`packages/flutter_tools/lib/src/features.dart`的配置。

```bash
/// The [Feature] for flutter web.
const Feature flutterWebFeature = Feature(
  name: 'Flutter for web',
  configSetting: 'enable-web',
  environmentOverride: 'FLUTTER_WEB',
  master: FeatureChannelSetting(
    available: true,
    enabledByDefault: false,
  ),
  dev: FeatureChannelSetting(
    available: true,
    enabledByDefault: false,
  ),
  beta: FeatureChannelSetting(
    available: true,
    enabledByDefault: false,
  ),
  stable: FeatureChannelSetting(
    available: true,
    enabledByDefault: false,
  ),
);
```

flutter 1.22.4 版本中可直接使用以下脚本替换：

```bash
sed -i '131s/beta/stable/g' flutter/packages/flutter_tools/lib/src/features.dart 
```

执行启动支持web ，并删除flutter tools snapshot

```bash
flutter config --enable-web
rm ~/flutter/bin/cache/flutter_tools.snapshot
```