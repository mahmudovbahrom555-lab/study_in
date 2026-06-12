ALTER TABLE topic_mastery
    DROP COLUMN IF EXISTS confidence_score,
    DROP COLUMN IF EXISTS consistency_score,
    DROP COLUMN IF EXISTS correct_streak;
