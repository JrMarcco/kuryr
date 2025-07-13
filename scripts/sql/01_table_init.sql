DROP TABLE IF EXISTS `biz_conf`;
CREATE TABLE `biz_conf` (
    `id`         BIGINT UNSIGNED NOT NULL PRIMARY KEY,
    `owner_id`   BIGINT UNSIGNED NOT NULL COMMENT '所有者ID',
    `owner_type` VARCHAR(64) NOT NULL COMMENT '所有者类型，例如：user, project, team...',
    `rate_limit` INT         NOT NULL DEFAULT -1 COMMENT '速率限制（请求/秒），-1表示不限制',
    `created_at` BIGINT UNSIGNED NOT NULL COMMENT '创建时间戳',
    `updated_at` BIGINT UNSIGNED NOT NULL COMMENT '更新时间戳',
    UNIQUE KEY `uk_owner` (`owner_id`, `owner_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务方配置表';
