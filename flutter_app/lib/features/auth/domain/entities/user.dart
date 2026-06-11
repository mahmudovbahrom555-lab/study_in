class User {
  const User({
    required this.id,
    required this.phone,
    required this.name,
    this.role,
    this.avatarUrl,
    required this.language,
    required this.isActive,
  });

  final String id;
  final String phone;
  final String name;
  final String? role;
  final String? avatarUrl;
  final String language;
  final bool isActive;

  bool get hasRole => role != null;
  bool get isTeacher => role == 'teacher';
  bool get isStudent => role == 'student';
  bool get isParent => role == 'parent';

  User copyWith({
    String? name,
    String? role,
    String? avatarUrl,
    String? language,
  }) {
    return User(
      id: id,
      phone: phone,
      name: name ?? this.name,
      role: role ?? this.role,
      avatarUrl: avatarUrl ?? this.avatarUrl,
      language: language ?? this.language,
      isActive: isActive,
    );
  }
}
