///
/// 用户信息
///
// ignore_for_file: argument_type_not_assignable
// ignore_for_file:  prefer_expression_function_bodies
import 'dart:convert';

class User {
  final String account;
  User({
    this.account,
  });

  User copyWith({
    String account,
  }) {
    return User(
      account: account ?? this.account,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'account': account,
    };
  }

  factory User.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return User(
      account: map['account'],
    );
  }

  String toJson() => json.encode(toMap());

  factory User.fromJson(String source) => User.fromMap(json.decode(source));

  @override
  String toString() => 'User(account: $account)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is User && o.account == account;
  }

  @override
  int get hashCode => account.hashCode;
}
