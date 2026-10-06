import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/theme/app_theme.dart';
import '../domain/ride_service.dart';

IconData rideServiceIcon(String token) => switch (token) {
  'car' => Icons.directions_car_outlined,
  'car-front' => Icons.directions_car_filled,
  _ => Icons.local_taxi,
};

class RiderServiceCards extends StatelessWidget {
  const RiderServiceCards({
    super.key,
    required this.scrollController,
    required this.scrollEnabled,
    required this.catalog,
    required this.onSelected,
    required this.onRetry,
  });

  final ScrollController scrollController;
  final bool scrollEnabled;
  final AsyncValue<List<RideService>> catalog;
  final ValueChanged<RideService> onSelected;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return ListView(
      controller: scrollController,
      physics: scrollEnabled
          ? const ClampingScrollPhysics()
          : const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(
          'Choose your ride',
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: AppSpacing.sm),
        ...catalog.when<List<Widget>>(
          skipLoadingOnRefresh: false,
          loading: () => const [
            ListTile(
              title: Text('Loading ride services...'),
              trailing: SizedBox.square(
                dimension: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            ),
          ],
          error: (error, stack) => [
            const Text('Unable to load ride services.'),
            DashboardPanelControl(
              child: TextButton(onPressed: onRetry, child: const Text('Retry')),
            ),
          ],
          data: (services) {
            if (services.isEmpty) {
              return [
                const Text('No ride services currently available.'),
                DashboardPanelControl(
                  child: TextButton(
                    onPressed: onRetry,
                    child: const Text('Retry'),
                  ),
                ),
              ];
            }
            return services
                .map((service) {
                  return Padding(
                    padding: const EdgeInsets.only(bottom: AppSpacing.sm),
                    child: DashboardPanelControl(
                      child: Card.outlined(
                        child: InkWell(
                          key: ValueKey('ride-service-card-${service.code}'),
                          borderRadius: BorderRadius.circular(AppRadii.lg),
                          onTap: () => onSelected(service),
                          child: Padding(
                            padding: const EdgeInsets.all(AppSpacing.md),
                            child: Row(
                              children: [
                                Icon(
                                  rideServiceIcon(service.presentationToken),
                                  size: 32,
                                ),
                                const SizedBox(width: AppSpacing.md),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        service.displayName,
                                        style: Theme.of(context)
                                            .textTheme
                                            .titleMedium,
                                      ),
                                      if (service.description.isNotEmpty) ...[
                                        const SizedBox(height: AppSpacing.xxs),
                                        Text(service.description),
                                      ],
                                    ],
                                  ),
                                ),
                                const SizedBox(width: AppSpacing.xs),
                                const Icon(Icons.chevron_right),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                  );
                })
                .toList(growable: false);
          },
        ),
      ],
    );
  }
}
