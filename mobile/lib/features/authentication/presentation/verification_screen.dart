import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/providers.dart';

class VerificationRequest {
  const VerificationRequest({
    required this.email,
    required this.verificationId,
  });

  final String email;
  final String verificationId;
}

class VerificationScreen extends ConsumerStatefulWidget {
  const VerificationScreen({super.key, required this.request});
  final VerificationRequest request;

  @override
  ConsumerState<VerificationScreen> createState() => _VerificationScreenState();
}

class _VerificationScreenState extends ConsumerState<VerificationScreen> {
  final _code = TextEditingController();
  late String _verificationId;

  @override
  void initState() {
    super.initState();
    _verificationId = widget.request.verificationId;
  }

  @override
  void dispose() {
    _code.dispose();
    super.dispose();
  }

  Future<void> _verify() async {
    if (_code.text.trim().isEmpty) return;
    final ok = await ref
        .read(sessionControllerProvider)
        .verify(_verificationId, _code.text);
    if (ok && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Email verified. Sign in to continue.')),
      );
      context.go('/login');
    }
  }

  Future<void> _resend() async {
    final challenge = await ref
        .read(sessionControllerProvider)
        .startVerification(widget.request.email);
    if (challenge != null && mounted) {
      setState(() {
        _verificationId = challenge;
        _code.clear();
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('A new verification code was sent.')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(sessionControllerProvider).state;
    return Scaffold(
      appBar: AppBar(title: const Text('Verify email')),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 420),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  'Enter the verification code sent to ${widget.request.email}.',
                ),
                const SizedBox(height: 16),
                TextField(
                  key: const Key('verificationCodeField'),
                  controller: _code,
                  keyboardType: TextInputType.number,
                  autofillHints: const [AutofillHints.oneTimeCode],
                  decoration: const InputDecoration(
                    labelText: 'Verification code',
                    border: OutlineInputBorder(),
                  ),
                ),
                if (state.error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    state.error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ],
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: state.busy ? null : _verify,
                  child: const Text('Verify'),
                ),
                TextButton(
                  key: const Key('resendVerificationButton'),
                  onPressed: state.busy ? null : _resend,
                  child: const Text('Send a new code'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
