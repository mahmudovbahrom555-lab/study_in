import 'package:flutter/material.dart';

import '../../domain/entities/group.dart';
import '../../../../core/localization/l10n.dart';

class MemberTile extends StatelessWidget {
  const MemberTile({
    super.key,
    required this.member,
    required this.isOwner,
    this.onRemove,
    this.onPaymentTap,
  });

  final GroupMember member;
  final bool isOwner;
  final VoidCallback? onRemove;
  final void Function(PaymentStatus)? onPaymentTap;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: CircleAvatar(
        backgroundImage: member.avatarUrl != null
            ? NetworkImage(member.avatarUrl!)
            : null,
        child: member.avatarUrl == null
            ? Text(member.name.substring(0, 1).toUpperCase())
            : null,
      ),
      title: Text(member.name),
      subtitle: Text(member.phone),
      trailing: isOwner
          ? Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _PaymentChip(
                  status: member.paymentStatus,
                  onTap: onPaymentTap,
                ),
                if (onRemove != null)
                  IconButton(
                    tooltip: context.l10n.removeFromGroup,
                    icon: const Icon(Icons.person_remove_outlined),
                    onPressed: onRemove,
                  ),
              ],
            )
          : _PaymentChip(status: member.paymentStatus),
    );
  }
}

class _PaymentChip extends StatelessWidget {
  const _PaymentChip({required this.status, this.onTap});

  final PaymentStatus status;
  final void Function(PaymentStatus)? onTap;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (status) {
      PaymentStatus.paid => (context.l10n.paymentPaid, Colors.green),
      PaymentStatus.pending => (context.l10n.paymentPending, Colors.orange),
      PaymentStatus.trial => (context.l10n.paymentTrial, Colors.grey),
    };

    final chip = Chip(
      label: Text(label, style: const TextStyle(fontSize: 12)),
      backgroundColor: color.withOpacity(0.15),
      side: BorderSide(color: color.withOpacity(0.4)),
      visualDensity: VisualDensity.compact,
    );

    if (onTap == null) return chip;

    // GestureDetector не сообщает экранному диктору, что это кнопка.
    return Semantics(
      button: true,
      label: context.l10n.paymentStatusSemantic(label),
      hint: context.l10n.changePaymentStatus,
      excludeSemantics: true,
      onTap: () => _showPicker(context),
      child: GestureDetector(
        onTap: () => _showPicker(context),
        child: chip,
      ),
    );
  }

  void _showPicker(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            for (final s in PaymentStatus.values)
              ListTile(
                title: Text(_label(context.l10n, s)),
                leading: Icon(
                  s == status ? Icons.radio_button_checked : Icons.radio_button_unchecked,
                ),
                onTap: () {
                  Navigator.pop(context);
                  onTap?.call(s);
                },
              ),
          ],
        ),
      ),
    );
  }

  String _label(AppLocalizations l10n, PaymentStatus s) => switch (s) {
        PaymentStatus.paid => l10n.paymentPaid,
        PaymentStatus.pending => l10n.paymentPendingLong,
        PaymentStatus.trial => l10n.paymentTrialLong,
      };
}
