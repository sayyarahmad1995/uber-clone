import 'dart:async';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_controller.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_panels.dart';

class FlowFake implements RideFlowRepository {
  Json ride = {'id': 'ride', 'status': 'requested', 'trip': null};
  Json? trip;
  int reads = 0;
  int actions = 0;
  bool loseResponse = false;
  Completer<void>? blocked;
  @override
  Future<Json> get(String path) async {
    reads++;
    if (blocked != null) await blocked!.future;
    if (path == '/v1/driver/trip') {
      if (trip == null) throw const ApiException('missing', 'No trip', statusCode: 404);
      return trip!;
    }
    if (path.endsWith('/offers')) return {'offers': [{'driver_user_id':'driver','selectable':true}]};
    if (path.endsWith('/driver-location')) throw const ApiException('missing','No location', statusCode:404);
    if (path.endsWith('/trips')) return {'trips': []};
    if (path.contains('/marketplace/')) return {'ride_requests': []};
    return ride;
  }
  @override
  Future<void> act(String path, {Json? data, bool put = false}) async {
    actions++;
    ride = {'id': 'ride', 'status': 'requested', 'trip': {'status':'assigned'}};
    if (loseResponse) throw const ApiException('network_error', 'Response lost');
  }
}
void main() {
  test('lost acceptance response reloads authoritative assignment without claiming success', () async {
    final repo = FlowFake()..loseResponse=true;
    final flow = RideFlowController(repo,rideId:'ride');
    await Future<void>.delayed(Duration.zero);
    expect(flow.offers, hasLength(1));
    await flow.act('/v1/ride-requests/ride/offers/driver/accept');
    expect(flow.status,'assigned');
    expect(flow.offers,isEmpty);
    expect(flow.error,contains('Response lost'));
    expect(flow.location,isNull);
    flow.dispose();
  });
  test('restores assigned Driver trip even while availability is offline', () async {
    final repo = FlowFake()..trip={'ride_request_id':'ride','status':'in_progress'};
    final flow = RideFlowController(repo);
    await Future<void>.delayed(Duration.zero);
    expect(flow.status,'in_progress');
    expect(flow.requests,isEmpty);
    flow.dispose();
  });
  test('background stops polling; dispose tolerates in-flight reads', () async {
    final repo = FlowFake();
    final flow = RideFlowController(repo,rideId:'ride',interval:const Duration(milliseconds:10));
    await Future<void>.delayed(Duration.zero);
    flow.setForeground(false);
    final reads=repo.reads;
    await Future<void>.delayed(const Duration(milliseconds:30));
    expect(repo.reads,reads);
    repo.blocked=Completer<void>();
    flow.setForeground(true);
    flow.dispose();
    repo.blocked!.complete();
    await Future<void>.delayed(Duration.zero);
  });
  test('commands cannot overlap an in-flight reload', () async {
    final repo = FlowFake()..blocked=Completer<void>();
    final flow = RideFlowController(repo,rideId:'ride');
    await flow.act('/accept');
    expect(repo.actions,0);
    flow.dispose();
    repo.blocked!.complete();
    await Future<void>.delayed(Duration.zero);
  });
  test('fare parsing is exact and rejects non-finite or excess precision input', () {
    expect(parseFareMinor('100.01'),10001);
    for(final input in ['NaN','Infinity','1.001','-1','0','1e4']) { expect(parseFareMinor(input),isNull); }
  });
  test('stale or future locations are never shown as live', () {
    expect(freshDriverLocation({'updated_at':DateTime.now().subtract(const Duration(minutes:3)).toIso8601String()}),isNull);
    expect(freshDriverLocation({'updated_at':DateTime.now().add(const Duration(minutes:3)).toIso8601String()}),isNull);
    expect(freshDriverLocation({'updated_at':DateTime.now().toIso8601String()}),isNotNull);
  });
}
