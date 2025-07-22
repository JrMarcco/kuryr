DROP TABLE IF EXISTS biz_config;
CREATE TABLE biz_config (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    channel_config JSONB NOT NULL,
    quota_config JSONB NOT NULL,
    callback_config JSONB NOT NULL,
    rate_limit INT NOT NULL DEFAULT -1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- 在 owner_id 列上创建索引
CREATE INDEX idx_biz_config_owner_id ON biz_config (owner_id);

COMMENT ON TABLE biz_config IS '用户信息表';
COMMENT ON COLUMN biz_config.owner_id IS '所有者 id ( biz_info.id )';
COMMENT ON COLUMN biz_config.channel_config IS '渠道配置';
COMMENT ON COLUMN biz_config.quota_config IS '配额配置';
COMMENT ON COLUMN biz_config.callback_config IS '回调配置';
COMMENT ON COLUMN biz_config.rate_limit IS '限流阈值 ( requests/s )，-1 表示不限制';
COMMENT ON COLUMN biz_config.created_at IS '创建时间戳（Unix 毫秒值）';
COMMENT ON COLUMN biz_config.updated_at IS '更新时间戳（Unix 毫秒值）';
