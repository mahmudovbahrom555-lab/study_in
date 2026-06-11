-- Этап 1: схема авторизации и пользователей.

-- ============================================
-- USERS
-- ============================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone           VARCHAR(20) UNIQUE NOT NULL,
    name            VARCHAR(100) NOT NULL DEFAULT '',
    role            VARCHAR(20) CHECK (role IN ('teacher', 'student', 'parent')),
    avatar_url      TEXT,
    language        VARCHAR(10) NOT NULL DEFAULT 'uz'
                        CHECK (language IN ('uz', 'uz-cyrl', 'ru', 'en')),
    organization_id UUID,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_phone ON users(phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role  ON users(role)  WHERE deleted_at IS NULL;

-- ============================================
-- SMS VERIFICATION CODES
-- ============================================

CREATE TABLE verification_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone      VARCHAR(20) NOT NULL,
    code       VARCHAR(6) NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_phone ON verification_codes(phone, created_at DESC);

-- ============================================
-- REFRESH TOKENS
-- ============================================

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    device_info VARCHAR(255),
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_user      ON refresh_tokens(user_id)     WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_token_hash ON refresh_tokens(token_hash) WHERE revoked_at IS NULL;

-- ============================================
-- FCM DEVICE TOKENS
-- ============================================

CREATE TABLE device_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL,
    platform   VARCHAR(20) CHECK (platform IN ('ios', 'android')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (token)
);

CREATE INDEX idx_device_tokens_user ON device_tokens(user_id);
