abstract interface class MapTiles {
  String get urlTemplate;
  String get attribution;
  String get userAgentPackageName;
}

class OpenStreetMapTiles implements MapTiles {
  const OpenStreetMapTiles();
  @override
  String get urlTemplate => 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
  @override
  String get attribution => '© OpenStreetMap contributors';
  @override
  String get userAgentPackageName => 'com.sayyarahmad.uberclone.uber_clone';
}
