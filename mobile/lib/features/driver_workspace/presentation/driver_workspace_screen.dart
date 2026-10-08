import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/models/account.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../../ride_flow/ride_flow_panels.dart';
import '../../ride_flow/domain/trip.dart';
import '../application/driver_onboarding_controller.dart';
import '../application/driver_route_controller.dart';
import '../domain/driver_onboarding.dart';
import '../domain/driver_profile.dart';
import 'operating_selection.dart';
import 'vehicle_service_application_section.dart';

class DriverWorkspaceScreen extends ConsumerStatefulWidget {
  const DriverWorkspaceScreen({super.key, required this.accountID});

  final String accountID;

  @override
  ConsumerState<DriverWorkspaceScreen> createState() =>
      _DriverWorkspaceScreenState();
}

class _DriverWorkspaceScreenState extends ConsumerState<DriverWorkspaceScreen>
    with WidgetsBindingObserver {
  final _mapController = RideMapController();
  bool _reapplying = false;
  String? _lastFittedRouteKey;
  String? _navigationInFlight;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    final driver = ref.read(driverControllerProvider);
    final flow = ref.read(driverTripControllerProvider);
    // Inactive can be a permission dialog; only actual background states end presence.
    if (state == AppLifecycleState.resumed) {
      driver.setForeground(true);
      if (driver.profile != null) {
        flow.setForeground(true);
        ref.read(driverRouteControllerProvider).retry();
      }
    } else if (state == AppLifecycleState.paused ||
        state == AppLifecycleState.hidden ||
        state == AppLifecycleState.detached) {
      driver.setForeground(false);
      if (driver.profile != null) {
        flow.setForeground(false);
      }
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  Future<void> _focusCurrentLocation({bool showError = true}) async {
    final controller = ref.read(driverControllerProvider);
    await controller.publishLocation();
    if (!mounted) {
      return;
    }
    final location = controller.location;
    if (location == null) {
      if (showError) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              controller.error ?? 'Unable to publish Driver location.',
            ),
          ),
        );
      }
      return;
    }
    await _mapController.move(
      RideMapPoint(location.latitude, location.longitude),
      15,
    );
  }

  @override
  Widget build(BuildContext context) {
    final driver = ref.watch(driverControllerProvider);
    final profile = driver.profile;
    final location = driver.location;
    final trip = profile == null
        ? null
        : ref.watch(driverTripControllerProvider).trip;
    if (profile != null) {
      ref.listen(driverTripControllerProvider, (_, flow) {
        if (flow.loaded && !flow.busy) {
          driver.setActiveTrip(flow.trip != null);
        }
      });
    }
    final route = profile == null
        ? const DriverRouteState()
        : ref.watch(driverRouteControllerProvider).state;
    _scheduleRouteFit(route, trip?.rideRequestId);
    final navigationTarget = _navigationTarget(trip);
    final navigationKey = _navigationKey(trip);
    final onboarding = driver.loaded && profile == null
        ? ref.watch(driverOnboardingControllerProvider)
        : null;
    final application = onboarding?.application;
    final loading =
        !driver.loaded ||
        (profile == null && onboarding != null && !onboarding.loaded);
    final onboardingFailed =
        !loading &&
        profile == null &&
        onboarding != null &&
        onboarding.error != null &&
        onboarding.services.isEmpty &&
        application == null;
    final setup =
        !loading &&
        !onboardingFailed &&
        profile == null &&
        (application == null || _reapplying);

    return RideDashboardScaffold(
      panelIdentity: loading
          ? 'driver-loading'
          : onboardingFailed
          ? 'driver-onboarding-error'
          : profile != null
          ? 'driver-readiness'
          : setup
          ? 'driver-onboarding-form'
          : 'driver-onboarding-${application!.status}',
      map: RideMap(
        mapController: _mapController,
        markers: [
          if (location != null)
            RideMapMarker(
              point: RideMapPoint(location.latitude, location.longitude),
              color: AppColors.success,
              label: 'Your published location',
            ),
          if (trip?.pickup != null)
            RideMapMarker(
              point: RideMapPoint(
                trip!.pickup!.latitude,
                trip.pickup!.longitude,
              ),
              color: AppColors.danger,
              label: 'pickup',
            ),
          if (trip?.destination != null)
            RideMapMarker(
              point: RideMapPoint(
                trip!.destination!.latitude,
                trip.destination!.longitude,
              ),
              color: AppColors.danger,
              label: 'destination',
            ),
        ],
        polylines: route.preview == null
            ? const []
            : [
                RideMapPolyline(
                  points: route.preview!.points
                      .map(
                        (point) =>
                            RideMapPoint(point.latitude, point.longitude),
                      )
                      .toList(growable: false),
                  color: Theme.of(context).colorScheme.primary,
                  width: 6,
                ),
              ],
      ),
      mapControls: profile == null
          ? null
          : Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                _DriverMapFocusButton(onPressed: _focusCurrentLocation),
                if (navigationTarget != null) ...[
                  const SizedBox(height: AppSpacing.sm),
                  FloatingActionButton.extended(
                    key: const Key('driverNavigationButton'),
                    heroTag: 'driver-navigation',
                    onPressed: _navigationInFlight == navigationKey
                        ? null
                        : _navigate,
                    icon: const Icon(Icons.navigation),
                    label: Text(
                      trip!.status == 'assigned'
                          ? 'Navigate to pickup'
                          : 'Navigate to destination',
                    ),
                  ),
                ],
                if (route.error != null) ...[
                  const SizedBox(height: AppSpacing.sm),
                  FloatingActionButton.small(
                    key: const Key('driverRouteRetryButton'),
                    heroTag: 'driver-route-retry',
                    tooltip: 'Retry driving route',
                    onPressed: () async {
                      await ref.read(driverTripControllerProvider).refresh();
                      if (!mounted) return;
                      await ref.read(driverRouteControllerProvider).retry();
                    },
                    child: const Icon(Icons.refresh),
                  ),
                ],
              ],
            ),
      floatingStatus: DashboardStatusCard(
        icon: Icons.local_taxi,
        title: 'Driver dashboard',
        message: route.loading
            ? route.status == 'assigned'
                  ? 'Loading route to pickup.'
                  : 'Loading route to destination.'
            : route.error ??
                  (route.preview != null
                      ? _routeEstimate(route)
                      : _statusMessage(
                          loading: loading,
                          onboardingFailed: onboardingFailed,
                          profile: profile,
                          application: application,
                          onlinePresenceReady: driver.onlinePresenceReady,
                        )),
      ),
      panelBuilder: (_, scrollController, scrollEnabled) {
        final physics = scrollEnabled
            ? const ClampingScrollPhysics()
            : const NeverScrollableScrollPhysics();

        if (loading || onboardingFailed) {
          return _LoadingPanel(
            scrollController: scrollController,
            physics: physics,
            busy: driver.busy || (onboarding?.busy ?? false),
            error: driver.error ?? onboarding?.error,
            onRetry: () {
              if (!driver.loaded || driver.error != null) {
                driver.load();
              }
              onboarding?.load();
            },
          );
        }

        if (profile != null) {
          return _DriverReadinessPanel(
            profile: profile,
            onlinePresenceReady: driver.onlinePresenceReady,
            location: location,
            busy: driver.busy,
            error: driver.error,
            scrollController: scrollController,
            physics: physics,
            onAvailabilityChanged: driver.setOnline,
            selectionValid: driver.operation?.valid == true,
            activeTrip: trip != null,
            onPublishLocation: _focusCurrentLocation,
            onRefresh: driver.load,
          );
        }

        if (!setup && application != null && onboarding != null) {
          return _DriverApplicationStatusPanel(
            application: application,
            busy: onboarding.busy,
            error: onboarding.error,
            scrollController: scrollController,
            physics: physics,
            onRefresh: () async {
              await onboarding.load();
              await driver.load();
            },
            onReapply: application.isRejected
                ? () => setState(() => _reapplying = true)
                : null,
          );
        }

        return _DriverSetupForm(
          key: ValueKey('setup-${widget.accountID}'),
          services: onboarding!.services,
          busy: onboarding.busy,
          error: onboarding.error,
          scrollController: scrollController,
          physics: physics,
          onReview: (name, service, vehicle) => _reviewAndSubmit(
            onboarding: onboarding,
            name: name,
            service: service,
            vehicle: vehicle,
          ),
          onCancel: _reapplying
              ? () => setState(() => _reapplying = false)
              : () async {
                  final result = await ref
                      .read(sessionControllerProvider)
                      .selectCapability(Capability.rider);
                  if (!context.mounted) return;
                  if (result.selected) {
                    context.go('/rider');
                  } else if (result.message != null) {
                    ScaffoldMessenger.of(context)
                        .showSnackBar(SnackBar(content: Text(result.message!)));
                  }
                },
          cancelLabel: _reapplying ? 'Cancel new application' : 'Back to Rider',
        );
      },
    );
  }


  RideMapPoint? _navigationTarget(TripSnapshot? trip) {
    if (trip?.rideRequestId == null) return null;
    final target = switch (trip!.status) {
      'assigned' => trip.pickup,
      'in_progress' => trip.destination,
      _ => null,
    };
    if (target == null ||
        !target.latitude.isFinite ||
        !target.longitude.isFinite ||
        target.latitude.abs() > 90 ||
        target.longitude.abs() > 180) {
      return null;
    }
    return RideMapPoint(target.latitude, target.longitude);
  }

  String? _navigationKey(TripSnapshot? trip) {
    final target = _navigationTarget(trip);
    if (target == null) return null;
    return '${widget.accountID}|${trip!.rideRequestId}|${trip.status}|'
        '${target.latitude},${target.longitude}';
  }

  Future<void> _navigate() async {
    final trip = ref.read(driverTripControllerProvider).trip;
    final target = _navigationTarget(trip);
    final key = _navigationKey(trip);
    if (target == null || key == null || _navigationInFlight == key) return;
    setState(() => _navigationInFlight = key);
    var opened = false;
    try {
      opened = await ref.read(drivingNavigationProvider).open(target);
    } catch (_) {
      // Keep a launcher failure recoverable without changing the Trip.
    }
    if (!mounted) return;
    if (_navigationInFlight == key) {
      setState(() => _navigationInFlight = null);
    }
    if (!opened &&
        _navigationKey(ref.read(driverTripControllerProvider).trip) == key) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Unable to open Google Maps. Try again.'),
        ),
      );
    }
  }

  String _routeEstimate(DriverRouteState route) {
    final preview = route.preview!;
    final kilometres = (preview.distanceMeters / 1000).toStringAsFixed(1);
    final minutes = (preview.durationSeconds / 60).ceil();
    final target = route.status == 'assigned' ? 'pickup' : 'destination';
    return 'Estimated route to $target: $kilometres km, $minutes min';
  }

  void _scheduleRouteFit(DriverRouteState route, String? rideRequestId) {
    final preview = route.preview;
    if (preview == null) {
      if (route.status == null) _lastFittedRouteKey = null;
      return;
    }
    final key = '$rideRequestId|${route.status}|${preview.encodedPolyline}';
    if (_lastFittedRouteKey == key) return;
    _lastFittedRouteKey = key;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || _lastFittedRouteKey != key) return;
      unawaited(
        _mapController.fit(
          preview.points.map(
            (point) => RideMapPoint(point.latitude, point.longitude),
          ),
        ),
      );
    });
  }

  String _statusMessage({
    required bool loading,
    required bool onboardingFailed,
    required DriverProfile? profile,
    required DriverOnboardingApplication? application,
    required bool onlinePresenceReady,
  }) {
    if (loading) return 'Loading Driver status.';
    if (onboardingFailed) return 'Unable to load Driver onboarding.';
    if (profile != null) {
      if (profile.isOnline && !onlinePresenceReady) {
        return 'Online presence unavailable. Go offline or retry.';
      }
      return profile.isOnline
          ? 'You are online. Keep your location updated.'
          : 'You are offline.';
    }
    if (_reapplying || application == null) {
      return 'Choose a service and submit your vehicle for review.';
    }
    if (application.isPending) {
      return 'Your Driver application is under review.';
    }
    if (application.isRejected) return 'Your Driver application was rejected.';
    return 'Your Driver application is approved.';
  }

  Future<void> _reviewAndSubmit({
    required DriverOnboardingController onboarding,
    required String name,
    required DriverServiceOption service,
    required DriverVehicle vehicle,
  }) async {
    final precheck = await onboarding.precheck(
      displayName: name,
      serviceCode: service.code,
      vehicle: vehicle,
    );
    if (!mounted || precheck == null) return;
    if (!precheck.eligible) {
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Vehicle not eligible'),
          content: Text(
            precheck.reasons.isEmpty
                ? 'This vehicle does not meet the selected service requirements.'
                : precheck.reasons.join('\n'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Change application'),
            ),
          ],
        ),
      );
      return;
    }

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Review application'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Driver: ${name.trim()}'),
            Text('Service: ${service.displayName}'),
            Text(
              'Vehicle: ${vehicle.make} ${vehicle.model} ${vehicle.modelYear ?? ''}',
            ),
            Text('Color: ${vehicle.color}'),
            Text('License plate: ${vehicle.licensePlate}'),
            const SizedBox(height: AppSpacing.sm),
            const Text(
              'Submitting sends these details for review. They do not become approved information immediately.',
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Back'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Submit for review'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    await onboarding.submit(
      displayName: name,
      serviceCode: service.code,
      vehicle: vehicle,
    );
    if (mounted && onboarding.error == null) {
      setState(() => _reapplying = false);
    }
  }
}

class _DriverMapFocusButton extends StatelessWidget {
  const _DriverMapFocusButton({required this.onPressed});

  final Future<void> Function() onPressed;

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton.small(
      key: const Key('driverCurrentLocationButton'),
      heroTag: 'driver-current-location',
      tooltip: 'Center map on your location',
      onPressed: () => onPressed(),
      child: const Icon(Icons.my_location),
    );
  }
}

class _LoadingPanel extends StatelessWidget {
  const _LoadingPanel({
    required this.scrollController,
    required this.physics,
    required this.busy,
    required this.error,
    required this.onRetry,
  });

  final ScrollController scrollController;
  final ScrollPhysics physics;
  final bool busy;
  final String? error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => ListView(
    controller: scrollController,
    physics: physics,
    padding: const EdgeInsets.all(AppSpacing.md),
    children: [
      Text(busy ? 'Loading Driver status...' : 'Unable to load Driver status'),
      if (error != null) Text(error!),
      DashboardPanelControl(
        child: TextButton(
          onPressed: busy ? null : onRetry,
          child: const Text('Retry'),
        ),
      ),
    ],
  );
}

class _DriverApplicationStatusPanel extends StatelessWidget {
  const _DriverApplicationStatusPanel({
    required this.application,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onRefresh,
    this.onReapply,
  });

  final DriverOnboardingApplication application;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final VoidCallback onRefresh;
  final VoidCallback? onReapply;

  @override
  Widget build(BuildContext context) {
    final title = application.isPending
        ? 'Application under review'
        : application.isRejected
        ? 'Application rejected'
        : 'Application approved';
    return ListView(
      controller: scrollController,
      physics: physics,
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(title, style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: AppSpacing.sm),
        Text('Service: ${application.service.displayName}'),
        Text(
          '${application.vehicle.make} ${application.vehicle.model} ${application.vehicle.modelYear ?? ''} - ${application.vehicle.color}',
        ),
        Text(application.vehicle.licensePlate),
        Text('Submitted ${application.submittedAt.toLocal()}'),
        if (application.isPending) ...[
          const SizedBox(height: AppSpacing.sm),
          const Text(
            'Your submitted details remain pending until the review decision is recorded.',
          ),
        ],
        if (application.isRejected) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(
            application.rejectionReason ??
                'The selected vehicle/service application was not approved.',
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        if (application.isApproved) ...[
          const SizedBox(height: AppSpacing.sm),
          const Text(
            'Approval is recorded. Operational Driver activation is handled by the approved vehicle/service foundation.',
          ),
        ],
        if (error != null)
          Text(
            error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        DashboardPanelControl(
          child: TextButton(
            onPressed: busy ? null : onRefresh,
            child: const Text('Refresh application'),
          ),
        ),
        if (onReapply != null)
          DashboardPanelControl(
            child: FilledButton(
              onPressed: busy ? null : onReapply,
              child: const Text('Apply again'),
            ),
          ),
      ],
    );
  }
}

class _DriverReadinessPanel extends StatelessWidget {
  const _DriverReadinessPanel({
    required this.profile,
    required this.onlinePresenceReady,
    required this.location,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onAvailabilityChanged,
    required this.selectionValid,
    required this.activeTrip,
    required this.onPublishLocation,
    required this.onRefresh,
  });

  final DriverProfile profile;
  final bool onlinePresenceReady;
  final bool selectionValid;
  final bool activeTrip;
  final PublishedDriverLocation? location;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final Future<void> Function(bool) onAvailabilityChanged;
  final Future<void> Function() onPublishLocation;
  final Future<void> Function() onRefresh;

  @override
  Widget build(BuildContext context) => ListView(
    controller: scrollController,
    physics: physics,
    padding: const EdgeInsets.all(AppSpacing.md),
    children: [
      Text(
        activeTrip
            ? 'Your active trip'
            : profile.isOnline && !onlinePresenceReady
            ? 'Online presence unavailable'
            : profile.isOnline
            ? 'You are online'
            : selectionValid
            ? 'Ready to go online'
            : 'Choose your operating vehicle',
        style: Theme.of(context).textTheme.headlineSmall,
      ),
      const SizedBox(height: AppSpacing.sm),
      Text(
        profile.displayName?.isNotEmpty == true
            ? profile.displayName!
            : 'Driver profile',
      ),
      if (activeTrip) const RideFlowPanel.driver(),
      if (!activeTrip) const OperatingSelectionControl(),
      if (!activeTrip) ...[
        const SizedBox(height: AppSpacing.md),
        const VehicleServiceApplicationSection(),
      ],
      const SizedBox(height: AppSpacing.md),
      DashboardPanelControl(
        child: FilledButton.icon(
          onPressed:
              busy || (!profile.isOnline && (!selectionValid || activeTrip))
              ? null
              : () => onAvailabilityChanged(!profile.isOnline),
          icon: const Icon(Icons.power_settings_new),
          label: Text(profile.isOnline ? 'Go offline' : 'Go online'),
        ),
      ),
      if (!activeTrip) const RideFlowPanel.driver(),
      const SizedBox(height: AppSpacing.sm),
      Text(
        location == null
            ? 'No location published during this visit.'
            : 'Location published at ${location!.updatedAt.toLocal()}',
      ),
      DashboardPanelControl(
        child: OutlinedButton.icon(
          onPressed: busy ? null : onPublishLocation,
          icon: const Icon(Icons.my_location),
          label: const Text('Update current location'),
        ),
      ),
      if (busy) const LinearProgressIndicator(),
      if (error != null)
        Text(
          error!,
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      DashboardPanelControl(
        child: TextButton(
          onPressed: busy ? null : onRefresh,
          child: const Text('Refresh status'),
        ),
      ),
    ],
  );
}

class _DriverSetupForm extends StatefulWidget {
  const _DriverSetupForm({
    super.key,
    required this.services,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onReview,
    this.onCancel,
    this.cancelLabel = 'Cancel new application',
  });

  final List<DriverServiceOption> services;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final Future<void> Function(String, DriverServiceOption, DriverVehicle)
  onReview;
  final VoidCallback? onCancel;
  final String cancelLabel;

  @override
  State<_DriverSetupForm> createState() => _DriverSetupFormState();
}

class _DriverSetupFormState extends State<_DriverSetupForm> {
  final _form = GlobalKey<FormState>();
  late final List<TextEditingController> _fields;
  String? _serviceCode;

  static const _labels = [
    'Display name',
    'Vehicle make',
    'Vehicle model',
    'Model year',
    'Vehicle color',
    'License plate',
  ];

  @override
  void initState() {
    super.initState();
    _fields = List.generate(6, (_) => TextEditingController());
    if (widget.services.isNotEmpty) {
      _serviceCode = widget.services.first.code;
    }
  }

  @override
  void didUpdateWidget(covariant _DriverSetupForm oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_serviceCode == null && widget.services.isNotEmpty) {
      _serviceCode = widget.services.first.code;
    }
  }

  @override
  void dispose() {
    for (final field in _fields) {
      field.dispose();
    }
    super.dispose();
  }

  DriverServiceOption? get _selectedService {
    for (final service in widget.services) {
      if (service.code == _serviceCode) return service;
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    final selectedService = _selectedService;
    return Form(
      key: _form,
      child: ListView(
        controller: widget.scrollController,
        physics: widget.physics,
        padding: const EdgeInsets.all(AppSpacing.md),
        children: [
          Text(
            'Become a Driver',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const Text('Choose one service for your first vehicle application.'),
          const SizedBox(height: AppSpacing.sm),
          if (widget.services.isEmpty)
            const Text('No Driver services are currently available.'),
          DashboardPanelControl(
            child: DropdownButtonFormField<String>(
              key: const Key('driver-service-field'),
              initialValue: _serviceCode,
              decoration: const InputDecoration(labelText: 'Service'),
              items: widget.services
                  .map(
                    (service) => DropdownMenuItem(
                      value: service.code,
                      child: Text(service.displayName),
                    ),
                  )
                  .toList(growable: false),
              onChanged: widget.busy || widget.services.isEmpty
                  ? null
                  : (value) => setState(() => _serviceCode = value),
              validator: (value) => value == null ? 'Select a service' : null,
            ),
          ),
          if (selectedService != null) ...[
            const SizedBox(height: AppSpacing.sm),
            Text(selectedService.description),
            if (selectedService.minimumModelYear != null)
              Text('Minimum model year: ${selectedService.minimumModelYear}'),
          ],
          for (var index = 0; index < _labels.length; index++)
            Padding(
              padding: const EdgeInsets.only(top: AppSpacing.sm),
              child: DashboardPanelControl(
                child: TextFormField(
                  key: ValueKey('driver-field-$index'),
                  controller: _fields[index],
                  enabled: !widget.busy,
                  decoration: InputDecoration(labelText: _labels[index]),
                  keyboardType: index == 3
                      ? TextInputType.number
                      : TextInputType.text,
                  validator: (value) {
                    if (value == null || value.trim().isEmpty) {
                      return 'Required';
                    }
                    if (index == 3) {
                      final year = int.tryParse(value.trim());
                      if (year == null ||
                          year < 1886 ||
                          year > DateTime.now().toUtc().year + 1) {
                        return 'Enter a valid model year';
                      }
                    }
                    return null;
                  },
                ),
              ),
            ),
          if (widget.error != null)
            Text(
              widget.error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          DashboardPanelControl(
            child: FilledButton(
              onPressed: widget.busy || selectedService == null
                  ? null
                  : () {
                      if (!_form.currentState!.validate()) return;
                      final values = _fields
                          .map((field) => field.text.trim())
                          .toList();
                      widget.onReview(
                        values[0],
                        selectedService,
                        DriverVehicle(
                          make: values[1],
                          model: values[2],
                          modelYear: int.parse(values[3]),
                          color: values[4],
                          licensePlate: values[5],
                        ),
                      );
                    },
              child: Text(widget.busy ? 'Checking...' : 'Review application'),
            ),
          ),
          if (widget.onCancel != null)
            DashboardPanelControl(
              child: TextButton(
                onPressed: widget.busy ? null : widget.onCancel,
                child: Text(widget.cancelLabel),
              ),
            ),
        ],
      ),
    );
  }
}
