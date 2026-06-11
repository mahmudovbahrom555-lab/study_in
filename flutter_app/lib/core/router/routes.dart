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
}
