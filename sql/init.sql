-- Create user_profiles table with only IDs
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id INT NOT NULL,
    profile_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id),
    PRIMARY KEY (user_id, profile_id)
);

-- Insert default profiles if not exist
INSERT INTO profiles (name) 
VALUES ('root'), ('admin'), ('guest'), ('users'), ('performance'), ('logs')
ON CONFLICT (name) DO NOTHING;
