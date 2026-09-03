# Driver public presentation

Rider offer comparison needs Driver and vehicle information that belongs to the
application, not to the authentication provider. This slice adds application-owned
Driver `display_name` and vehicle `model_year` and exposes them through the existing
Rider offer-comparison response.

## Ownership boundary

- `display_name` is public marketplace presentation data owned by the Driver domain.
  Marketplace reads must not fetch it from Ory Kratos or expose identity-provider traits.
- Vehicle make/model/model year/color are public pre-assignment presentation fields.
- License plate remains private before assignment.
- Existing Driver rows are preserved by nullable migration columns. Re-onboarding fills
  the new fields; new onboarding requires a nonblank display name and a plausible model year.
- Driver operational eligibility is not coupled to presentation completeness during this
  migration slice, so legacy Driver accounts are not silently disabled.

## Photos

Driver and vehicle photos remain required product presentation, but they are not stored
as arbitrary external URLs in this slice. The repository has no application-owned media
upload/storage boundary yet. A later focused media slice should define that boundary and
then add photo references to Driver and vehicle presentation without coupling marketplace
logic to a specific storage vendor.

## Rider response

`GET /v1/ride-requests/{ride_request_id}/offers` may include:

```json
{
  "driver": {"display_name": "Sayyar Ahmad"},
  "vehicle": {
    "make": "Toyota",
    "model": "Corolla",
    "model_year": 2024,
    "color": "White"
  }
}
```

For legacy rows, `driver` may be null and `model_year` may be absent/unknown until the
Driver updates onboarding data. Existing fare, distance, selectability, privacy, and
assignment semantics are unchanged.
