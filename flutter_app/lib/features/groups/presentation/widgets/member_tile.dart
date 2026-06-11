import 'package:flutter/material.dart';

import '../../domain/entities/group.dart';

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
      PaymentStatus.paid => ('Оплачено', Colors.green),
      PaymentStatus.pending => ('Ожидает', Colors.orange),
      PaymentStatus.trial => ('Пробный', Colors.grey),
    };

    final chip = Chip(
      label: Text(label, style: const TextStyle(fontSize: 12)),
      backgroundColor: color.withOpacity(0.15),
      side: BorderSide(color: color.withOpacity(0.4)),
      visualDensity: VisualDensity.compact,
    );

    if (onTap == null) return chip;

    return GestureDetector(
      onTap: () => _showPicker(context),
      child: chip,
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
                title: Text(_label(s)),
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

  String _label(PaymentStatus s) => switch (s) {
        PaymentStatus.paid => 'Оплачено',
        PaymentStatus.pending => 'Ожидает оплаты',
        PaymentStatus.trial => 'Пробный период',
      };
}
