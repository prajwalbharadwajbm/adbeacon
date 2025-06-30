-- Create campaigns table
CREATE TABLE campaigns (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_url VARCHAR(512) NOT NULL,
    cta VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create targeting_rules table
CREATE TABLE targeting_rules (
    id BIGSERIAL PRIMARY KEY,
    campaign_id VARCHAR(255) NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    dimension VARCHAR(20) NOT NULL CHECK (dimension IN ('country', 'os', 'app')),
    rule_type VARCHAR(20) NOT NULL CHECK (rule_type IN ('include', 'exclude')),
    values TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_status_updated ON campaigns(status, updated_at);
CREATE INDEX idx_targeting_rules_campaign_id ON targeting_rules(campaign_id);
CREATE INDEX idx_targeting_rules_dimension ON targeting_rules(dimension);
CREATE INDEX idx_targeting_rules_campaign_dimension ON targeting_rules(campaign_id, dimension);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for campaigns table
CREATE TRIGGER update_campaigns_updated_at 
    BEFORE UPDATE ON campaigns 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO campaigns (id, name, image_url, cta, status) VALUES
('spotify', 'Spotify - Music for everyone', 'https://somelink', 'Download', 'ACTIVE'),
('duolingo', 'Duolingo: Best way to learn', 'https://somelink2', 'Install', 'ACTIVE'),
('subwaysurfer', 'Subway Surfer', 'https://somelink3', 'Play', 'ACTIVE');

-- Insert targeting rules
INSERT INTO targeting_rules (campaign_id, dimension, rule_type, values) VALUES
('spotify', 'country', 'include', ARRAY['US', 'Canada']),
('duolingo', 'os', 'include', ARRAY['Android', 'iOS']),
('duolingo', 'country', 'exclude', ARRAY['US']),
('subwaysurfer', 'os', 'include', ARRAY['Android']),
('subwaysurfer', 'app', 'include', ARRAY['com.gametion.ludokinggame']); 