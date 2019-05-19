CREATE TABLE `members` (
  `uid` char(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'cpw-platform members uuid',
  `permission` smallint(6) NOT NULL COMMENT '權限值',
  `is_blockade` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否被封鎖(1:是,0:否)',
  `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '建立時間',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;