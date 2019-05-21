CREATE TABLE `rooms` (
  `room_id` char(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '房間id',
  `is_message` tinyint(1) NOT NULL COMMENT '是否可以聊天',
  `is_bonus` tinyint(1) NOT NULL COMMENT '是否可發/搶紅包',
  `is_follow` tinyint(1) NOT NULL COMMENT '是否可發/跟注',
  `day_limit` tinyint(4) NOT NULL COMMENT '聊天限制天數範圍',
  `amount_limit` int(11) NOT NULL COMMENT '儲值金額聊天限制',
  `dml_limit` int(11) NOT NULL COMMENT '打碼量聊天限制',
  PRIMARY KEY (`room_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;