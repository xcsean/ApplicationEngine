SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for t_global_config
-- ----------------------------
DROP TABLE IF EXISTS `t_global_config`;
CREATE TABLE `t_global_config` (
  `t_category` varchar(127) NOT NULL,
  `t_key` varchar(127) NOT NULL,
  `t_value` varchar(255) NOT NULL,
  PRIMARY KEY (`t_category`,`t_key`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;

-- ----------------------------
-- Table structure for t_protocol
-- ----------------------------
DROP TABLE IF EXISTS `t_protocol`;
CREATE TABLE `t_protocol` (
  `proto_id` int(10) NOT NULL,
  `player_limit_enable` int(10) NOT NULL DEFAULT '0',
  `player_limit_count` int(10) NOT NULL DEFAULT '0',
  `player_limit_duration` int(10) NOT NULL DEFAULT '0',
  `server_limit_enable` int(10) NOT NULL DEFAULT '0',
  `server_limit_count` int(10) NOT NULL DEFAULT '0',
  `server_limit_duration` int(10) NOT NULL DEFAULT '0',
  PRIMARY KEY (`proto_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for t_server
-- ----------------------------
DROP TABLE IF EXISTS `t_server`;
CREATE TABLE `t_server` (
  `app` varchar(32) NOT NULL COMMENT 'for example app',
  `server` varchar(32) NOT NULL COMMENT 'for example gobby',
  `division` varchar(32) NOT NULL COMMENT 'for example 101',
  `node` varchar(128) NOT NULL COMMENT 'for example app.gobby.101',
  `use_agent` int(255) NOT NULL DEFAULT '1' COMMENT 'use agent to monitor',
  `node_status` int(255) NOT NULL DEFAULT '0' COMMENT 'process status, 0 is ready',
  `service_status` int(255) NOT NULL DEFAULT '0' COMMENT 'service in process status, 0 is ready',
  PRIMARY KEY (`app`,`server`,`division`,`node`) USING BTREE
) ENGINE=MyISAM DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for t_service
-- ----------------------------
DROP TABLE IF EXISTS `t_service`;
CREATE TABLE `t_service` (
  `app` varchar(32) NOT NULL COMMENT 'for example app',
  `server` varchar(32) NOT NULL COMMENT 'for example gobby',
  `division` varchar(32) NOT NULL COMMENT 'for example 101',
  `node` varchar(128) NOT NULL COMMENT 'for example app.gobby.101',
  `service` varchar(64) NOT NULL COMMENT 'for example app.gobby.notify',
  `service_ip` varchar(256) NOT NULL COMMENT 'intranet ip, such as 192.168.1.10',
  `service_port` int(10) NOT NULL COMMENT '17000',
  `admin_port` int(10) NOT NULL COMMENT '17001',
  `rpc_port` int(10) NOT NULL COMMENT '17002',
  PRIMARY KEY (`app`,`server`,`division`,`node`,`service`) USING BTREE
) ENGINE=MyISAM DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
