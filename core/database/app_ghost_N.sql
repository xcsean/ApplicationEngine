SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for t_userasset
-- ----------------------------
DROP TABLE IF EXISTS `t_userasset`;
CREATE TABLE `t_userasset`  (
  `ghostid` int(10) UNSIGNED NOT NULL COMMENT 'ghost id',
  `uuid` bigint(20) UNSIGNED NOT NULL COMMENT 'user unique id',
  `lockerid` bigint(20) UNSIGNED NOT NULL COMMENT 'locker id',
  `expiredtime` bigint(20) NOT NULL COMMENT 'expired time',
  `revision` bigint(20) UNSIGNED NOT NULL COMMENT 'revision',
  `asset` mediumblob COMMENT 'asset data',
  PRIMARY KEY (`ghostid`, `uuid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;

SET FOREIGN_KEY_CHECKS=1;
