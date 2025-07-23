DROP TABLE IF EXISTS biz_config;
CREATE TABLE biz_config (
    id BIGINT PRIMARY KEY,
    channel_config JSONB NOT NULL,
    quota_config JSONB NOT NULL,
    callback_config JSONB NOT NULL,
    rate_limit INT NOT NULL DEFAULT -1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

COMMENT ON TABLE biz_config IS '业务方配置信息表，与 biz_info 一对一关系';
COMMENT ON COLUMN biz_config.id IS 'id ( 与 biz_info.id 保持一致 )';
COMMENT ON COLUMN biz_config.channel_config IS '渠道配置';
COMMENT ON COLUMN biz_config.quota_config IS '配额配置';
COMMENT ON COLUMN biz_config.callback_config IS '回调配置';
COMMENT ON COLUMN biz_config.rate_limit IS '限流阈值 ( requests/s )，0 表示不限制';
COMMENT ON COLUMN biz_config.created_at IS '创建时间戳（Unix 毫秒值）';
COMMENT ON COLUMN biz_config.updated_at IS '更新时间戳（Unix 毫秒值）';
