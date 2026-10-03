import 'ride_request.dart';

enum RiderPlaceField { pickup, destination }

class PlaceSuggestion {
  const PlaceSuggestion({required this.placeId, required this.label});

  final String placeId;
  final String label;

  factory PlaceSuggestion.fromJson(Map<String, dynamic> json) =>
      PlaceSuggestion(
        placeId: json['place_id'] as String,
        label: json['label'] as String,
      );
}

class PlaceSelection {
  const PlaceSelection({
    this.placeId,
    required this.label,
    required this.point,
  });

  final String? placeId;
  final String label;
  final GeoPoint point;

  factory PlaceSelection.fromJson(Map<String, dynamic> json) => PlaceSelection(
    placeId: json['place_id'] as String?,
    label: json['label'] as String,
    point: GeoPoint(
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
    ),
  );
}
