DROP TYPE IF EXISTS active_status_enum;
CREATE TYPE active_status_enum AS ENUM ('active', 'inactive');

DROP TYPE IF EXISTS audit_status_enum;
CREATE TYPE audit_status_enum AS ENUM ('pending', 'auditing', 'approved', 'rejected');

DROP TYPE IF EXISTS channel_enum;
CREATE TYPE channel_enum AS ENUM ('1', '2');
