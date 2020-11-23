# pike admin


Enable web feature

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