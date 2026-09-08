-- atlas:txmode file
-- VOC-1445: VOC-031-D06 permits exactly the canonical "en" app language at
-- launch. Keep legacy direct-write values readable during rollout, while every
-- new write must match the only language the product can render.

ALTER TABLE user_settings
  DROP CONSTRAINT user_settings_app_language_valid,
  ADD CONSTRAINT user_settings_app_language_valid
    CHECK (app_language = 'en') NOT VALID;
