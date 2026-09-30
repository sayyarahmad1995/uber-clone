-- Existing Economy/Comfort remain available to Riders. New catalog entries
-- start unpublished so operations can approve supply before Rider exposure.
ALTER TABLE driver_service_catalog
    ADD COLUMN rider_visible BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN presentation_token TEXT NOT NULL DEFAULT 'car',
    ADD CONSTRAINT driver_service_catalog_presentation_not_blank
        CHECK (BTRIM(presentation_token) <> '');

UPDATE driver_service_catalog
SET rider_visible = TRUE,
    presentation_token = CASE code
        WHEN 'comfort' THEN 'car-front'
        ELSE 'car'
    END
WHERE code IN ('economy', 'comfort');
