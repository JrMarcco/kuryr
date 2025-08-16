-- 业务信息表
DROP TABLE IF EXISTS biz_info;
CREATE TABLE biz_info (
    id BIGSERIAL PRIMARY KEY,
    biz_type biz_type_enum NOT NULL,
    biz_key VARCHAR(64) NOT NULL,
    biz_secret VARCHAR(128) NOT NULL,
    biz_name VARCHAR(128) NOT NULL,
    contact varchar(64) NOT NULL,
    contact_email varchar(128) NOT NULL,
    creator_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT uk_biz_key UNIQUE (biz_key)
);

COMMENT ON TABLE biz_info IS '业务信息表';
COMMENT ON COLUMN biz_info.id IS 'id';
COMMENT ON COLUMN biz_info.biz_name IS '业务名';
COMMENT ON COLUMN biz_info.biz_type IS '业务类型';
COMMENT ON COLUMN biz_info.biz_key IS '业务 key 用于识别业务方身份';
COMMENT ON COLUMN biz_info.biz_secret IS '业务密钥 用于认证';
COMMENT ON COLUMN biz_info.contact IS '业务联系人';
COMMENT ON COLUMN biz_info.contact_email IS '联系人邮箱';
COMMENT ON COLUMN biz_info.creator_id IS '创建人 id';
COMMENT ON COLUMN biz_info.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN biz_info.updated_at IS '更新时间戳 ( Unix 毫秒值 )';

-- 字段索引：业务名
-- 查询场景：where biz_name like '%?%'
--      注意 gin 索引支持模糊匹配 where provider_name like '%keyword%'，
--      而 B-tree 索引只支持前缀匹配。
--
-- pg_trgm 为 postgresql 的官方拓展：
--      ├── pg_trgm 扩展 (Extension)
--      ├── 函数 (Functions)
--      │   ├── similarity()
--      │   ├── show_trgm()
--      │   ├── word_similarity()
--      │   └── ...
--      ├── 操作符 (Operators)
--      │   ├── % (相似性)
--      │   ├── <-> (距离)
--      │   ├── <<-> (左边界距离)
--      │   └── ...
--      └── 操作符类 (Operator Classes)
--          ├── gin_trgm_ops  -> GIN  索引用 ( 倒排索引：Generalized Inverted Index，擅长处理包含关系和精确查找 )
--          └── gist_trgm_ops -> GIST 索引用 ( 空间索引：Generalized Search Tree，基于 R-tree 等树状结构，擅长处理范围、距离、包含关系 )
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_biz_info_biz_name_gin ON biz_info USING gin(biz_name gin_trgm_ops);

-- 业务配置信息表
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

COMMENT ON TABLE biz_config IS '业务配置信息表';
COMMENT ON COLUMN biz_config.id IS 'id ( 与 biz_info.id 保持一致 )';
COMMENT ON COLUMN biz_config.owner_type IS '业务类型';
COMMENT ON COLUMN biz_config.channel_config IS '渠道配置';
COMMENT ON COLUMN biz_config.quota_config IS '配额配置';
COMMENT ON COLUMN biz_config.callback_config IS '回调配置';
COMMENT ON COLUMN biz_config.rate_limit IS '限流阈值 ( requests/s ) 0 表示不限制';
COMMENT ON COLUMN biz_config.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN biz_config.updated_at IS '更新时间戳 ( Unix 毫秒值 )';

-- 供应商信息表
DROP TABLE IF EXISTS provider_info;
CREATE TABLE provider_info (
    id BIGSERIAL PRIMARY KEY,
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
COMMENT ON COLUMN provider_info.region_id IS '区域 id';
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

-- 字段索引：供应商名称
-- 查询场景：where provider_name like '?%' / where provider_name = ?
-- CREATE INDEX idx_provider_info_provider_name ON provider_info(provider_name);

-- 字段索引：渠道
-- 查询场景：where channel = ?
CREATE INDEX idx_provider_info_channel ON provider_info(channel);

-- 渠道模板信息表
DROP TABLE IF EXISTS channel_template;
CREATE TABLE channel_template (
    id BIGSERIAL PRIMARY KEY,
    owner_id VARCHAR(128) NOT NULL,
    owner_type VARCHAR(16) NOT NULL,
    tpl_name VARCHAR(128) NOT NULL,
    tpl_desc VARCHAR(128) NOT NULL,
    channel channel_enum NOT NULL,
    notification_type notification_type_enum NOT NULL,
    activated_version_id BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

COMMENT ON TABLE channel_template IS '渠道模板信息表';
COMMENT ON COLUMN channel_template.id IS 'id';
COMMENT ON COLUMN channel_template.owner_id IS '所属业务 id';
COMMENT ON COLUMN channel_template.owner_type IS '所属业务类型';
COMMENT ON COLUMN channel_template.tpl_name IS '模板名';
COMMENT ON COLUMN channel_template.tpl_desc IS '模板描述';
COMMENT ON COLUMN channel_template.channel IS '渠道';
COMMENT ON COLUMN channel_template.notification_type IS '消息类型';
COMMENT ON COLUMN channel_template.activated_version_id IS '当前激活版本 id';
COMMENT ON COLUMN channel_template.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN channel_template.updated_at IS '更新时间戳 ( Unix 毫秒值 )';

-- 组合索引：所属业务 + 消息类型
-- 查询场景：where owner_id = ? / where owner_id = ? and notification_type = ?
CREATE INDEX idx_channel_template_owner_notification_type ON channel_template(owner_id, notification_type);

-- 渠道模板版本信息表
DROP TABLE IF EXISTS channel_template_version;
CREATE TABLE channel_template_version (
    id BIGSERIAL PRIMARY KEY,
    tpl_id BIGINT NOT NULL,
    version_name VARCHAR(128) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    context TEXT NOT NULL,
    apply_remark VARCHAR(128) NOT NULL,
    audit_id BIGINT NOT NULL,
    auditor_id BIGINT NOT NULL,
    audit_time BIGINT NOT NULL DEFAULT 0,
    audit_status audit_status_enum NOT NULL,
    rejection_reason VARCHAR(128) NOT NULL,
    last_review_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

COMMENT ON TABLE channel_template_version IS '渠道模板版本信息表';
COMMENT ON COLUMN channel_template_version.id IS 'id';
COMMENT ON COLUMN channel_template_version.tpl_id IS '模板 id';
COMMENT ON COLUMN channel_template_version.version_name IS '版本名';
COMMENT ON COLUMN channel_template_version.signature IS '签名信息';
COMMENT ON COLUMN channel_template_version.context IS '模板内容';
COMMENT ON COLUMN channel_template_version.apply_remark IS '申请备注信息';
COMMENT ON COLUMN channel_template_version.audit_id IS '审批记录 id';
COMMENT ON COLUMN channel_template_version.auditor_id IS '审批人 id';
COMMENT ON COLUMN channel_template_version.audit_time IS '审批时间';
COMMENT ON COLUMN channel_template_version.audit_status IS '审批状态';
COMMENT ON COLUMN channel_template_version.rejection_reason IS '审批拒绝理由';
COMMENT ON COLUMN channel_template_version.last_review_at IS '上次审查时间';
COMMENT ON COLUMN channel_template_version.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN channel_template_version.updated_at IS '更新时间戳 ( Unix 毫秒值 )';

-- 组合索引：模板 id + 审核状态
-- 查询场景： where tpl_id = ? / where tpl_id = ? and audit_status = ?
CREATE INDEX idx_channel_template_version_tpl_audit ON channel_template_version(tpl_id, audit_status);

-- 渠道模板供应商信息表
DROP TABLE IF EXISTS channel_template_provider;
CREATE TABLE channel_template_provider (
    id BIGSERIAL PRIMARY KEY,
    tpl_id BIGINT NOT NULL,
    tpl_version_id BIGINT NOT NULL,
    provider_id BIGINT NOT NULL,
    provider_name VARCHAR(128) NOT NULL,
    provider_tpl_id VARCHAR(128) NOT NULL,
    provider_channel channel_enum NOT NULL,
    audit_request_id VARCHAR(64) NOT NULL,
    audit_status audit_status_enum NOT NULL,
    rejection_reason VARCHAR(128) NOT NULL,
    last_review_at BIGINT NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

COMMENT ON TABLE channel_template_provider IS '渠道模板供应商信息表';
COMMENT ON COLUMN channel_template_provider.id IS 'id';
COMMENT ON COLUMN channel_template_provider.tpl_id IS '模板 id';
COMMENT ON COLUMN channel_template_provider.tpl_version_id IS '模板版本 id';
COMMENT ON COLUMN channel_template_provider.provider_id IS '供应商 id';
COMMENT ON COLUMN channel_template_provider.provider_name IS '供应商名称';
COMMENT ON COLUMN channel_template_provider.provider_tpl_id IS '供应商侧模板 id';
COMMENT ON COLUMN channel_template_provider.provider_channel IS '供应商渠道';
COMMENT ON COLUMN channel_template_provider.audit_request_id IS '审批请求 id';
COMMENT ON COLUMN channel_template_provider.audit_status IS '审批状态';
COMMENT ON COLUMN channel_template_provider.rejection_reason IS '审批拒绝理由';
COMMENT ON COLUMN channel_template_provider.last_review_at IS '上次审查时间';
COMMENT ON COLUMN channel_template_provider.created_at IS '创建时间戳 ( Unix 毫秒值 )';
COMMENT ON COLUMN channel_template_provider.updated_at IS '更新时间戳 ( Unix 毫秒值 )';

-- 字段索引：版本 id
-- 查询场景：where tpl_version_id in (?)
CREATE INDEX idx_channel_template_tpl_version ON channel_template_provider(tpl_version_id);

-- 组合索引：模板 id + 版本 id
-- 查询场景：where tpl_id = ? and tpl_version_id in (?) / where tpl_id = ? and tpl_version_id = ?
CREATE INDEX idx_channel_template_provider_tpl_version ON channel_template_provider(tpl_id, tpl_version_id);
