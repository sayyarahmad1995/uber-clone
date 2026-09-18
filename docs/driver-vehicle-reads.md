# Driver vehicle and approved enrollment reads

`GET /v1/driver/vehicles` requires an authenticated account with Driver capability. Ownership comes from the authenticated user, never a client-supplied user ID. The response is `{ "vehicles": [...] }`; no records is HTTP 200 with an empty list.

Each persisted vehicle includes `id`, `make`, `model`, nullable `model_year`, `color`, `license_plate`, and `approved_service_enrollments`. Each enrollment includes `service_code`, `display_name`, `approved_at`, and `service_active`. Reviewer identities are not exposed.

Vehicles are ordered by creation time then ID. Enrollments are ordered by catalog sort order then service code. Only persisted enrollments are returned: Comfort does not implicitly enroll Economy. Inactive catalog services remain visible as approved enrollments with `service_active: false`.

The read works for approved Drivers before operational activation. Legacy vehicles without enrollments remain visible with an empty enrollment list; neither vehicle verification nor operating eligibility is inferred from their existence. Independent vehicle verification state is not yet represented by the schema.

The mobile Vehicles screen reads these records, displays all vehicles and their explicit approved services, and supports retry after a failed read. If there are no persisted vehicles, the onboarding application remains visible as a submitted historical snapshot. Driver Details uses the current operating vehicle selection and its active approved enrollments; it does not treat the onboarding service as a current selection.

Vehicle registration and editing remain separate from operating selection. Operating selection chooses a vehicle only; marketplace service eligibility is derived from that selected vehicle's active approved enrollments under ADR-0009.

Validation: run `go test ./...` and, with `TEST_DATABASE_URL` pointing to a dedicated database ending `_test`, `go test ./internal/driver -run TestPostgresRepository -count=1 -v`. In mobile run `dart format lib test` and `flutter test`.
