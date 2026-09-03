package offer

// Both marketplace read models use aliases rr (ride) and l (Driver location).
// Coordinates are validated on write. Clamp rounding error at antipodal points.
const pickupDistanceSQL = `(2 * 6371000.0 * ASIN(SQRT(LEAST(1.0, GREATEST(0.0,
	POWER(SIN(RADIANS(l.latitude - rr.pickup_latitude) / 2), 2) +
	COS(RADIANS(rr.pickup_latitude)) * COS(RADIANS(l.latitude)) *
	POWER(SIN(RADIANS(l.longitude - rr.pickup_longitude) / 2), 2)
)))))`
