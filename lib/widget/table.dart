///
/// Table组件
///
import 'package:flutter/material.dart';

import '../config/application.dart';

typedef OperationCallback = void Function(int);

final _defaultPadding = EdgeInsets.only(
  left: 2 * Application.defaultPadding,
  right: 2 * Application.defaultPadding,
  top: 1.5 * Application.defaultPadding,
  bottom: 1.5 * Application.defaultPadding,
);

// XTable table widget
class XTable extends StatelessWidget {
  final List<String> headers;
  final List<List<dynamic>> contents;
  final Map<String, double> columnWidths;

  const XTable(
    this.headers,
    this.contents, {
    this.columnWidths,
    Key key,
  })  : assert(
          headers != null,
          'Header can not be null',
        ),
        super(key: key);
  // _createRowItem create row item
  Widget _createRowItem(String text) => Padding(
        padding: _defaultPadding,
        child: Text(
          text ?? '--',
        ),
      );
  // _createRowListItem create row list item
  Widget _createRowListItem(List arr) => Padding(
        padding: _defaultPadding,
        child: Column(
          children: arr?.map((e) {
            Widget content;
            if (e is Widget) {
              content = e;
            } else {
              content = Text(e?.toString());
            }
            return Row(
              children: [
                Icon(
                  Icons.brightness_1,
                  size: 8,
                  color: Application.greyBlackColor,
                ),
                Container(
                  width: Application.defaultPadding / 2,
                ),
                content,
              ],
            );
          })?.toList(),
        ),
      );
  @override
  Widget build(BuildContext context) {
    final widths = <int, TableColumnWidth>{};
    // 根据配置转换为表格每列宽度限制
    if (columnWidths != null && columnWidths.isNotEmpty) {
      columnWidths.forEach((key, value) {
        var index = 0;
        var foundIndex = -1;
        headers.forEach((element) {
          if (element == key) {
            foundIndex = index;
          }
          index++;
        });
        if (foundIndex != -1) {
          widths[foundIndex] = FixedColumnWidth(value);
        }
      });
    }
    // 生成表头
    final headerItems = headers
        .map((e) => Container(
              color: Application.greyColor,
              padding: _defaultPadding,
              child: Text(
                e,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                ),
              ),
            ))
        .toList();
    final rows = <TableRow>[
      TableRow(
        children: headerItems,
      ),
    ];

    // 将内容添加至表格
    contents?.forEach((items) {
      final rowItems = items.map((element) {
        // 如果元素是widget，则直接返回
        if (element is Widget) {
          return element;
        }
        // 如果是列表，则以列表形式处理
        if (element is List) {
          return _createRowListItem(element);
        }
        return _createRowItem(element?.toString());
      }).toList();
      rows.add(TableRow(
        children: rowItems,
      ));
    });
    final borderSide = BorderSide(
      color: Application.defaultBorderColor,
    );
    return Table(
      columnWidths: widths,
      border: TableBorder(
        horizontalInside: borderSide,
        top: borderSide,
        bottom: borderSide,
      ),
      children: rows,
    );
  }
}

// XConfigTable config table widget
class XConfigTable extends StatelessWidget {
  final List<String> headers;
  final List<List<dynamic>> contents;
  final Map<String, double> columnWidths;
  final OperationCallback onUpdate;
  final OperationCallback onDelete;
  const XConfigTable({
    @required this.headers,
    @required this.contents,
    @required this.onUpdate,
    @required this.onDelete,
    this.columnWidths,
    Key key,
  })  : assert(
          onUpdate != null,
          'Update function can not be null',
        ),
        assert(
          onDelete != null,
          'Delete function can not be null',
        ),
        assert(
          headers != null,
          'Header can not be null',
        ),
        super(key: key);
  // _createOpertaionsRowItem create operation row items
  Widget _createOpertaionsRowItem(
          OperationCallback onUpdate, OperationCallback onDelete, int index) =>
      Padding(
        padding: EdgeInsets.only(
          left: 2 * Application.defaultPadding,
        ),
        child: Row(
          children: [
            IconButton(
              icon: Icon(
                Icons.edit,
                size: Application.normalFontSize,
              ),
              onPressed: () => onUpdate(index),
            ),
            IconButton(
              icon: Icon(
                Icons.delete_outline,
                size: Application.normalFontSize,
              ),
              onPressed: () => onDelete(index),
            ),
          ],
        ),
      );

  @override
  Widget build(BuildContext context) {
    headers.add('Operations');
    final newContents = <List<dynamic>>[];
    // 对列表中每一行增加opertaions的操作
    if (contents != null && contents.isNotEmpty) {
      var index = 0;
      contents.forEach((element) {
        final i = index;
        final newItems = [];
        newItems.addAll(element);
        newItems.add(_createOpertaionsRowItem(onUpdate, onDelete, i));
        index++;
        newContents.add(newItems);
      });
    }
    if (columnWidths != null) {
      columnWidths['Operations'] = 155;
    }
    return XTable(
      headers,
      newContents,
      columnWidths: columnWidths,
    );
  }
}
