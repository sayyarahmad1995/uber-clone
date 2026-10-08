/// Rider-visible service metadata. Service codes are backend-owned strings.
class RideService {
  const RideService({
    required this.code,
    required this.displayName,
    required this.description,
    required this.displayOrder,
    required this.presentationToken,
    this.pricingRequired = false,
  });

  final String code;
  final String displayName;
  final String description;
  final int displayOrder;
  final String presentationToken;
  final bool pricingRequired;

  factory RideService.fromJson(Map<String, dynamic> json) => RideService(
    code: json['code'] as String,
    displayName: json['display_name'] as String,
    description: json['description'] as String? ?? '',
    displayOrder: (json['display_order'] as num).toInt(),
    presentationToken: json['presentation_token'] as String? ?? 'car',
    pricingRequired: json['pricing_required'] as bool? ?? false,
  );
}
