class OperatingState {
  const OperatingState({
    this.vehicleId,
    this.valid = false,
    this.canChange = true,
  });
  final String? vehicleId;
  final bool valid;
  final bool canChange;
  factory OperatingState.fromJson(Map<String, dynamic> json) {
    final selection = json['selection'] as Map<String, dynamic>?;
    return OperatingState(
      vehicleId: selection?['vehicle_id'] as String?,
      valid: selection?['valid'] == true,
      canChange: json['can_change'] == true,
    );
  }
}
