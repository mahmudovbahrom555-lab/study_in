import 'dart:io';

import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/widgets/error_view.dart';
import '../../domain/entities/assignment.dart';
import '../../domain/entities/submission.dart';
import '../providers/assignments_provider.dart';
import '../widgets/attachment_views.dart';

/// Лимиты совпадают с бэкендом (DECISIONS.md, 2026-10-07).
const _maxFiles = 5;
const _photoMaxSide = 1600.0;
const _photoQuality = 80;

/// Экран сдачи ДЗ учеником: комментарий и до 5 файлов. Файлы грузятся сразу
/// после выбора, «Сдать» отправляет только их ID — сдача атомарная.
class SubmitWorkPage extends ConsumerStatefulWidget {
  const SubmitWorkPage({super.key, required this.assignment});

  final Assignment assignment;

  @override
  ConsumerState<SubmitWorkPage> createState() => _SubmitWorkPageState();
}

/// Вложение на экране: уже на сервере или ещё загружается с телефона.
class _Slot {
  _Slot.local(this.path, this.name);
  _Slot.uploaded(AttachedFile f)
      : path = null,
        name = f.name,
        file = f;

  final String? path;
  final String name;
  AttachedFile? file;
  double progress = 0;
  String? error;
  CancelToken? cancel;

  bool get isUploading => file == null && error == null;
  bool get isLocalImage =>
      path != null && RegExp(r'\.(jpe?g|png|heic)$', caseSensitive: false).hasMatch(path!);
}

class _SubmitWorkPageState extends ConsumerState<SubmitWorkPage> {
  final _comment = TextEditingController();
  final _slots = <_Slot>[];
  Submission? _existing;
  Object? _loadError;
  bool _loading = true;
  bool _sending = false;

  Assignment get _a => widget.assignment;
  bool get _readOnly => _existing?.isGraded ?? false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    for (final s in _slots) {
      s.cancel?.cancel();
    }
    _comment.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _loadError = null;
    });
    try {
      final sub = await ref
          .read(assignmentsRepositoryProvider)
          .mySubmission(_a.groupId, _a.id);
      if (!mounted) return;
      setState(() {
        _existing = sub;
        _comment.text = sub?.comment ?? '';
        _slots
          ..clear()
          ..addAll((sub?.files ?? const []).map(_Slot.uploaded));
        _loading = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() {
          _loadError = e;
          _loading = false;
        });
      }
    }
  }

  int get _free => _maxFiles - _slots.length;

  Future<void> _fromCamera() async {
    final x = await ImagePicker().pickImage(
      source: ImageSource.camera,
      maxWidth: _photoMaxSide,
      maxHeight: _photoMaxSide,
      imageQuality: _photoQuality,
    );
    if (x != null) _add([_Slot.local(x.path, x.name)]);
  }

  Future<void> _fromGallery() async {
    final picker = ImagePicker();
    // Сжатие на телефоне: длинная сторона ≤1600 px, JPEG; HEIC тоже становится JPEG.
    final List<XFile> picked = _free > 1
        ? await picker.pickMultiImage(
            maxWidth: _photoMaxSide,
            maxHeight: _photoMaxSide,
            imageQuality: _photoQuality,
            limit: _free,
          )
        : [
            if (await picker.pickImage(
              source: ImageSource.gallery,
              maxWidth: _photoMaxSide,
              maxHeight: _photoMaxSide,
              imageQuality: _photoQuality,
            )
                case final x?)
              x,
          ];
    _add(picked.map((x) => _Slot.local(x.path, x.name)).toList());
  }

  Future<void> _fromFiles() async {
    final res = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: const ['pdf', 'doc', 'docx'],
      allowMultiple: _free > 1,
    );
    if (res == null) return;
    _add([
      for (final f in res.files)
        if (f.path != null) _Slot.local(f.path!, f.name),
    ]);
  }

  void _add(List<_Slot> slots) {
    if (!mounted || slots.isEmpty) return;
    final fitting = slots.take(_free).toList();
    setState(() => _slots.addAll(fitting));
    if (fitting.length < slots.length) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(context.l10n.errTooManyFiles)));
    }
    for (final s in fitting) {
      _upload(s);
    }
  }

  Future<void> _upload(_Slot s) async {
    setState(() {
      s.error = null;
      s.progress = 0;
      s.cancel = CancelToken();
    });
    try {
      final f = await ref.read(assignmentsRepositoryProvider).uploadFile(
            path: s.path!,
            name: s.name,
            cancelToken: s.cancel,
            onProgress: (sent, total) {
              if (mounted && total > 0) setState(() => s.progress = sent / total);
            },
          );
      if (mounted) setState(() => s.file = f);
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      if (mounted) setState(() => s.error = userErrorMessage(e));
    } catch (e) {
      if (mounted) setState(() => s.error = userErrorMessage(e));
    }
  }

  void _remove(_Slot s) {
    s.cancel?.cancel();
    setState(() => _slots.remove(s));
  }

  Future<void> _submit() async {
    final l10n = context.l10n;
    if (_slots.any((s) => s.isUploading)) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(l10n.hwWaitUploads)));
      return;
    }
    final comment = _comment.text.trim();
    final ids = [for (final s in _slots) if (s.file != null) s.file!.id];
    if (comment.isEmpty && ids.isEmpty) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(l10n.errSubmissionEmpty)));
      return;
    }
    setState(() => _sending = true);
    try {
      await ref.read(assignmentsRepositoryProvider).submit(
            _a.groupId,
            _a.id,
            comment: comment.isEmpty ? null : comment,
            fileIds: ids,
          );
      ref.read(assignmentsProvider(_a.groupId).notifier).load();
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(l10n.hwSubmittedOk)));
      Navigator.of(context).pop();
    } catch (e) {
      if (!mounted) return;
      setState(() => _sending = false);
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      appBar: AppBar(title: Text(_a.title)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _loadError != null
              ? ErrorView(error: _loadError, onRetry: _load)
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    _AssignmentHeader(assignment: _a),
                    if (_existing case final sub? when sub.isGraded) ...[
                      const SizedBox(height: 16),
                      _GradeCard(submission: sub),
                    ],
                    const SizedBox(height: 24),
                    Text(l10n.hwYourWork, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _comment,
                      readOnly: _readOnly,
                      minLines: 2,
                      maxLines: 5,
                      decoration: InputDecoration(labelText: l10n.commentOptional),
                    ),
                    const SizedBox(height: 12),
                    for (final s in _slots)
                      _SlotTile(
                        slot: s,
                        readOnly: _readOnly,
                        onRemove: () => _remove(s),
                        onRetry: () => _upload(s),
                      ),
                    if (!_readOnly && _free > 0) ...[
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          OutlinedButton.icon(
                            icon: const Icon(LucideIcons.camera, size: 18),
                            label: Text(l10n.hwCamera),
                            onPressed: _fromCamera,
                          ),
                          OutlinedButton.icon(
                            icon: const Icon(LucideIcons.images, size: 18),
                            label: Text(l10n.hwGallery),
                            onPressed: _fromGallery,
                          ),
                          OutlinedButton.icon(
                            icon: const Icon(LucideIcons.paperclip, size: 18),
                            label: Text(l10n.hwFile),
                            onPressed: _fromFiles,
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Text(
                        l10n.hwAttachHint,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ],
                ),
      bottomNavigationBar: _loading || _loadError != null || _readOnly
          ? null
          : SafeArea(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
                child: FilledButton(
                  onPressed: _sending ? null : _submit,
                  style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
                  child: _sending
                      ? const SizedBox.square(
                          dimension: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(_existing == null ? l10n.submit : l10n.hwResubmit),
                ),
              ),
            ),
    );
  }
}

class _AssignmentHeader extends StatelessWidget {
  const _AssignmentHeader({required this.assignment});

  final Assignment assignment;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final due = assignment.dueDate;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (assignment.description != null)
          Text(assignment.description!, style: theme.textTheme.bodyLarge),
        if (due != null) ...[
          const SizedBox(height: 8),
          Text(
            context.l10n.dueDate(DateFormat('dd.MM.yyyy HH:mm').format(due)),
            style: theme.textTheme.bodyMedium?.copyWith(
              color: assignment.isOverdue
                  ? theme.colorScheme.error
                  : theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ],
    );
  }
}

class _GradeCard extends StatelessWidget {
  const _GradeCard({required this.submission});

  final Submission submission;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.gradeValue('${submission.grade}'),
              style: theme.textTheme.titleLarge?.copyWith(color: theme.colorScheme.primary),
            ),
            if (submission.teacherNote case final note? when note.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(l10n.hwTeacherNote, style: theme.textTheme.labelMedium),
              Text(note),
            ],
          ],
        ),
      ),
    );
  }
}

class _SlotTile extends StatelessWidget {
  const _SlotTile({
    required this.slot,
    required this.readOnly,
    required this.onRemove,
    required this.onRetry,
  });

  final _Slot slot;
  final bool readOnly;
  final VoidCallback onRemove;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    final f = slot.file;
    final Widget? preview = slot.isLocalImage
        ? Image.file(File(slot.path!), fit: BoxFit.cover, width: 48, height: 48)
        : (f != null && f.isImage)
            ? networkPreview(f)
            : null;

    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: AttachmentThumb(child: preview),
      title: Text(slot.name, maxLines: 1, overflow: TextOverflow.ellipsis),
      subtitle: slot.error != null
          ? Text(slot.error!, style: TextStyle(color: theme.colorScheme.error))
          : slot.isUploading
              ? Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: LinearProgressIndicator(value: slot.progress > 0 ? slot.progress : null),
                )
              : null,
      onTap: f != null ? () => openAttachment(context, f) : null,
      trailing: readOnly
          ? null
          : Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (slot.error != null)
                  IconButton(
                    tooltip: l10n.retry,
                    icon: const Icon(LucideIcons.rotateCw),
                    onPressed: onRetry,
                  ),
                IconButton(
                  tooltip: l10n.hwRemoveFile,
                  icon: const Icon(LucideIcons.x),
                  onPressed: onRemove,
                ),
              ],
            ),
    );
  }
}
