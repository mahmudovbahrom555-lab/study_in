/// Куда ушёл код входа. У Telegram [delivered] = false, если номер ещё не
/// подтверждён в боте: приложение предлагает открыть [botUrl], и бот пришлёт
/// код, как только человек поделится номером.
class SendCodeResult {
  const SendCodeResult({
    required this.channel,
    required this.delivered,
    this.botUrl,
  });

  final String channel;
  final bool delivered;
  final String? botUrl;

  bool get viaTelegram => channel == 'telegram';
  bool get needsBot => viaTelegram && !delivered && botUrl != null;
}
