DROP TABLE IF EXISTS points_log;
DROP INDEX IF EXISTS uk_member_invite_code;
ALTER TABLE member
    DROP COLUMN IF EXISTS invited_by,
    DROP COLUMN IF EXISTS invite_code,
    DROP COLUMN IF EXISTS points;
DROP TABLE IF EXISTS member_address;
