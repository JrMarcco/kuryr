DROP TYPE IF EXISTS biz_type_enum;
CREATE TYPE biz_type_enum AS ENUM ('individual', 'organization');

DROP TYPE IF EXISTS active_status_enum;
CREATE TYPE active_status_enum AS ENUM ('active', 'inactive');

DROP TYPE IF EXISTS audit_status_enum;
CREATE TYPE audit_status_enum AS ENUM ('pending', 'auditing', 'approved', 'rejected');

DROP TYPE IF EXISTS channel_enum;
CREATE TYPE channel_enum AS ENUM ('1', '2');

DROP TYPE IF EXISTS notification_type_enum;
CREATE TYPE notification_type_enum AS ENUM ('1', '2')
