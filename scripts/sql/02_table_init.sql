DROP TABLE IF EXISTS biz_config;
CREATE TABLE biz_config (
    id BIGINT PRIMARY KEY,
    owner_type VARCHAR(16) NOT NULL,
    channel_config JSONB NOT NULL,
    quota_config JSONB NOT NULL,
    callback_config JSONB NOT NULL,
    rate_limit INT NOT NULL DEFAULT -1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

COMMENT ON TABLE biz_config IS '业务方配置信息表';
COMMENT ON COLUMN biz_config.id IS 'id ( 与 biz_info.id 保持一致 )';
COMMENT ON COLUMN biz_config.owner_type IS '所属业务类型';
COMMENT ON COLUMN biz_config.channel_config IS '渠道配置';
COMMENT ON COLUMN biz_config.quota_config IS '配额配置';
COMMENT ON COLUMN biz_config.callback_config IS '回调配置';
COMMENT ON COLUMN biz_config.rate_limit IS '限流阈值 ( requests/s ) 0 表示不限制';
COMMENT ON COLUMN biz_config.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN biz_config.updated_at IS '更新时间戳 ( Unix 毫秒值 )';


DROP TABLE IF EXISTS provider_info;
CREATE TABLE provider_info (
    id BIGINT PRIMARY KEY,
    provider_name VARCHAR(128) NOT NULL,
    channel channel_enum NOT NULL,
    endpoint VARCHAR(128) NOT NULL,
    region_id VARCHAR(128) NOT NULL,
    app_id VARCHAR(128) NOT NULL,
    api_key VARCHAR(128) NOT NULL,
    api_secret VARCHAR(128) NOT NULL,
    weight INT NOT NULL DEFAULT 0,
    qps_limit INT NOT NULL DEFAULT 0,
    daily_limit INT NOT NULL DEFAULT 0,
    audit_callback_url VARCHAR(128) NOT NULL,
    active_status active_status_enum NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

COMMENT ON TABLE provider_info IS '供应商信息表';
COMMENT ON COLUMN provider_info.id IS 'id';
COMMENT ON COLUMN provider_info.provider_name IS '供应商名称';
COMMENT ON COLUMN provider_info.channel IS '渠道';
COMMENT ON COLUMN provider_info.endpoint IS '接口地址';
COMMENT ON COLUMN provider_info.region_id IS '区域 ID';
COMMENT ON COLUMN provider_info.app_id IS '应用 ID';
COMMENT ON COLUMN provider_info.api_key IS '接口密钥';
COMMENT ON COLUMN provider_info.api_secret IS '接口密钥';
COMMENT ON COLUMN provider_info.weight IS '权重';
COMMENT ON COLUMN provider_info.qps_limit IS '每秒请求限制';
COMMENT ON COLUMN provider_info.daily_limit IS '每日请求限制';
COMMENT ON COLUMN provider_info.audit_callback_url IS '审核回调地址';
COMMENT ON COLUMN provider_info.active_status IS '状态';
COMMENT ON COLUMN provider_info.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN provider_info.updated_at IS '更新时间戳 ( Unix 毫秒值 )';
