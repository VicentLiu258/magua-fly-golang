-- SMS verification codes for phone-based auth flows
CREATE TYPE sms_scene AS ENUM ('register', 'login', 'reset_pwd', '2fa');

CREATE TABLE sms_verification_codes (
    id BIGSERIAL PRIMARY KEY,
    phone VARCHAR(32) NOT NULL,
    scene sms_scene NOT NULL,
    code VARCHAR(16) NOT NULL,
    ip VARCHAR(64),
    expired_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sms_verification_phone_scene_created_at
    ON sms_verification_codes(phone, scene, created_at DESC);

