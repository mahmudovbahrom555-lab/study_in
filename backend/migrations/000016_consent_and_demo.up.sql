-- consent_events: GDPR-compliant audit trail of user consent changes.
-- Each row is immutable — changes are new rows, not updates.
-- consent_version ties to the privacy policy version the user agreed to.
CREATE TABLE consent_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type    VARCHAR(50) NOT NULL,   -- data_sharing | anonymous_research | network_learning
    consent_version VARCHAR(20) NOT NULL DEFAULT 'v1.0',
    granted         BOOLEAN     NOT NULL,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consent_user_type ON consent_events(user_id, consent_type, created_at DESC);

-- groups.is_demo: marks groups created by the Demo seeder.
-- Demo groups return hardcoded rich insights so a new teacher sees the full
-- platform value immediately, before any real students join.
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_demo      BOOLEAN     NOT NULL DEFAULT FALSE;
-- groups.cefr_level: the target proficiency level for Course Coverage computation.
ALTER TABLE groups ADD COLUMN IF NOT EXISTS cefr_level   VARCHAR(5);
