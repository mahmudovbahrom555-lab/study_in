import '../../../core/localization/l10n.dart';
import '../domain/entities/attendance.dart';

/// Подпись статуса на языке интерфейса. Тексты живут в слое представления,
/// а не в доменной сущности.
extension AttendanceStatusL10n on AttendanceStatus {
  String label(AppLocalizations l10n) => switch (this) {
        AttendanceStatus.present => l10n.attPresent,
        AttendanceStatus.absent => l10n.attAbsent,
        AttendanceStatus.late => l10n.attLate,
        AttendanceStatus.excused => l10n.attExcused,
      };
}
