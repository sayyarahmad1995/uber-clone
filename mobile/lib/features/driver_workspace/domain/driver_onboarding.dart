import 'driver_profile.dart';

class DriverServiceOption {
  const DriverServiceOption({
    required this.code,
    required this.displayName,
    required this.description,
    this.minimumModelYear,
    this.impliedServiceCode,
  });

  final String code;
  final String displayName;
  final String description;
  final int? minimumModelYear;
  final String? impliedServiceCode;

  factory DriverServiceOption.fromJson(Map<String, dynamic> json) =>
      DriverServiceOption(
        code: json['code'] as String,
        displayName: json['display_name'] as String,
        description: json['description'] as String? ?? '',
        minimumModelYear: json['minimum_model_year'] as int?,
        impliedServiceCode: json['implied_service_code'] as String?,
      );
}

class DriverOnboardingPrecheck {
  const DriverOnboardingPrecheck({
    required this.eligible,
    required this.reasons,
    required this.service,
  });

  final bool eligible;
  final List<String> reasons;
  final DriverServiceOption service;

  factory DriverOnboardingPrecheck.fromJson(Map<String, dynamic> json) =>
      DriverOnboardingPrecheck(
        eligible: json['eligible'] as bool,
        reasons: (json['reasons'] as List<dynamic>? ?? const []).cast<String>(),
        service: DriverServiceOption.fromJson(
          json['service'] as Map<String, dynamic>,
        ),
      );
}

class DriverOnboardingApplication {
  const DriverOnboardingApplication({
    required this.id,
    required this.displayName,
    required this.status,
    required this.service,
    required this.vehicle,
    required this.submittedAt,
    this.rejectionReason,
    this.decidedAt,
  });

  final String id;
  final String displayName;
  final String status;
  final DriverServiceOption service;
  final DriverVehicle vehicle;
  final String? rejectionReason;
  final DateTime submittedAt;
  final DateTime? decidedAt;

  bool get isPending => status == 'pending';
  bool get isApproved => status == 'approved';
  bool get isRejected => status == 'rejected';

  factory DriverOnboardingApplication.fromJson(Map<String, dynamic> json) =>
      DriverOnboardingApplication(
        id: json['id'] as String,
        displayName: json['display_name'] as String,
        status: json['status'] as String,
        service: DriverServiceOption.fromJson(
          json['service'] as Map<String, dynamic>,
        ),
        vehicle: DriverVehicle.fromJson(
          json['vehicle'] as Map<String, dynamic>,
        ),
        rejectionReason: switch (json['rejection_reason']) {
          final String value when value.isNotEmpty => value,
          _ => null,
        },
        submittedAt: DateTime.parse(json['submitted_at'] as String),
        decidedAt: json['decided_at'] == null
            ? null
            : DateTime.parse(json['decided_at'] as String),
      );
}
