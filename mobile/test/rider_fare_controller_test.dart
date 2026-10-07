import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/rider_request/application/rider_fare_controller.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';
import 'package:uber_clone/features/rider_request/domain/route_preview.dart';
import 'package:uber_clone/features/rider_request/domain/ride_service.dart';
void main() {
 RoutePreview preview(int amount,String version)=>RoutePreview(distanceMeters:2000,durationSeconds:300,encodedPolyline:'??_ibE_ibE',suggestedFare:Money(amountMinor:amount,currency:'PKR'),pricingPolicyVersion:version);
 test('suggestion prefills an untouched proposal and preserves edits on refresh',(){
  final c=RiderFareController();c.invalidate('A');c.applyPreview('A',preview(12500,'v1'));
  expect(c.text,'125.00');expect(c.version,'v1');expect(c.userEdited,false);
  c.edit('1100');c.applyPreview('A',preview(22500,'v2'));
  expect(c.text,'1100');expect(c.version,'v2');expect(c.userEdited,true);c.dispose();
 });
 test('changed input clears ownership and ignores an older preview',(){
  final c=RiderFareController();c.invalidate('A');c.applyPreview('A',preview(12500,'v1'));c.edit('1100');
  c.invalidate('B');expect(c.text,isEmpty);expect(c.version,isNull);expect(c.userEdited,false);
  c.applyPreview('A',preview(99900,'old'));expect(c.text,isEmpty);
  c.applyPreview('B',preview(22500,'v2'));expect(c.text,'225.00');c.dispose();
 });
 test('route-only Driver responses remain valid and incomplete pricing is rejected',(){
  final wire={'route':{'distance_meters':2000,'duration_seconds':300,'encoded_polyline':'??_ibE_ibE'}};
  expect(RoutePreview.fromJson(wire).suggestedFare,isNull);
  expect(()=>RoutePreview.fromJson({...wire,'suggested_fare':{'amount_minor':12500,'currency':'PKR'}}),throwsFormatException);
  expect(()=>RoutePreview.fromJson({...wire,'suggested_fare':{'amount_minor':12500,'currency':'USD'},'pricing_policy_version':'v1'}),throwsFormatException);
 });
 test('catalog prices require an explicit server signal',(){
  final wire={'code':'synthetic','display_name':'Third','display_order':1,'presentation_token':'future'};
  expect(RideService.fromJson(wire).pricingRequired,false);
  expect(RideService.fromJson({...wire,'pricing_required':true}).pricingRequired,true);
 });
}
