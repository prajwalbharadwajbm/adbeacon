-- Drop triggers
DROP TRIGGER IF EXISTS update_campaigns_updated_at ON campaigns;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_targeting_rules_campaign_dimension;
DROP INDEX IF EXISTS idx_targeting_rules_dimension;
DROP INDEX IF EXISTS idx_targeting_rules_campaign_id;
DROP INDEX IF EXISTS idx_campaigns_status_updated;
DROP INDEX IF EXISTS idx_campaigns_status;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS targeting_rules;
DROP TABLE IF EXISTS campaigns; 