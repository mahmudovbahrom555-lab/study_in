-- Extend topic_mastery with confidence and consistency dimensions.
-- confidence_score: how certain the student is (response time proxy, initially from answer streak).
--   Range [0,1]. Starts at 0.5 (neutral). Rises on consecutive correct, falls on wrong.
-- consistency_score: how reliably the student performs across sessions (not just today).
--   Range [0,1]. Computed from variance of recent attempt outcomes.
-- Both fields feed the future AI Coach for personalised learning plans.
ALTER TABLE topic_mastery
    ADD COLUMN IF NOT EXISTS confidence_score  NUMERIC(4,3) NOT NULL DEFAULT 0.500,
    ADD COLUMN IF NOT EXISTS consistency_score NUMERIC(4,3) NOT NULL DEFAULT 0.500,
    ADD COLUMN IF NOT EXISTS correct_streak    INT          NOT NULL DEFAULT 0;

COMMENT ON COLUMN topic_mastery.confidence_score  IS 'EMA of per-answer confidence [0,1]. Updated on each RecordAnswer call.';
COMMENT ON COLUMN topic_mastery.consistency_score IS 'Variance-based score [0,1]. 1 = perfectly consistent. Updated in batch.';
COMMENT ON COLUMN topic_mastery.correct_streak    IS 'Current consecutive correct answers for this topic.';
