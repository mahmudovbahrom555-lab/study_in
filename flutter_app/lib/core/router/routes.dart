abstract class Routes {
  static const splash = '/';
  static const phone = '/phone';
  static const verify = '/verify';
  static const roleSelect = '/role-select';
  static const profileSetup = '/profile-setup';
  static const home = '/home';
  static const groups = '/groups';
  static String group(String id) => '/groups/$id';
  static String groupFeed(String id) => '/groups/$id/feed';
  static String groupQuizzes(String id) => '/groups/$id/quizzes';
  static String quiz(String groupId, String quizId) =>
      '/groups/$groupId/quizzes/$quizId';
  static String quizAttempt(String groupId, String quizId) =>
      '/groups/$groupId/quizzes/$quizId/attempt';
  static String groupGrades(String groupId) => '/groups/$groupId/grades';
  static String studentGrades(String groupId, String studentId) =>
      '/groups/$groupId/students/$studentId/grades';
  static String groupAttendance(String groupId) =>
      '/groups/$groupId/attendance';
  static String studentAttendance(String groupId, String studentId) =>
      '/groups/$groupId/students/$studentId/attendance';
  static String groupAssignments(String groupId) =>
      '/groups/$groupId/assignments';
  static const notifications = '/notifications';
  static const parentChildren = '/parent/children';
  static String aiInsights(String groupId) => '/groups/$groupId/ai-insights';
  static String studentProgress(String groupId, String studentId) =>
      '/groups/$groupId/students/$studentId/progress';
  static const ownerRisk = '/owner/risk-alerts';
  static String parentRoi(String studentId, String groupId) =>
      '/parent/children/$studentId/groups/$groupId/roi';
}
