import 'driver_profile.dart';

class ApprovedServiceEnrollment {
  const ApprovedServiceEnrollment({required this.serviceCode, required this.displayName, required this.approvedAt, required this.serviceActive});
  final String serviceCode;
  final String displayName;
  final DateTime approvedAt;
  final bool serviceActive;

  factory ApprovedServiceEnrollment.fromJson(Map<String, dynamic> json) => ApprovedServiceEnrollment(
    serviceCode: json['service_code'] as String,
    displayName: json['display_name'] as String,
    approvedAt: DateTime.parse(json['approved_at'] as String),
    serviceActive: json['service_active'] as bool,
  );
}

class RegisteredVehicle {
  const RegisteredVehicle({required this.id, required this.vehicle, required this.enrollments});
  final String id;
  final DriverVehicle vehicle;
  final List<ApprovedServiceEnrollment> enrollments;

  factory RegisteredVehicle.fromJson(Map<String, dynamic> json) => RegisteredVehicle(
    id: json['id'] as String,
    vehicle: DriverVehicle.fromJson(json),
    enrollments: (json['approved_service_enrollments'] as List)
        .map((item) => ApprovedServiceEnrollment.fromJson(item as Map<String, dynamic>)).toList(),
  );
}
