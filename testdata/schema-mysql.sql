# Dump of table actions
# ------------------------------------------------------------

DROP TABLE IF EXISTS `actions`;

CREATE TABLE `actions` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '事件ID',
  `user_id` int(11) NOT NULL COMMENT '用户',
  `content` text NOT NULL COMMENT '事件内容',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table articles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `articles`;

CREATE TABLE `articles` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `pid` int(11) NOT NULL COMMENT '文章父ID',
  `doc_id` int(11) NOT NULL COMMENT '所属文档ID',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '文章标题',
  `user_id` int(11) NOT NULL COMMENT '创建者',
  `last_user_id` int(11) NOT NULL COMMENT '最后编辑者',
  `content` longtext NOT NULL COMMENT '文章内容',
  `type` int(11) NOT NULL COMMENT '文章类型1文章，2外链，3引用，4分类，5 swagger',
  `link` varchar(255) NOT NULL DEFAULT '' COMMENT '链接地址，type=2时有效',
  `reference_id` int(11) NOT NULL DEFAULT '0' COMMENT '引用ID type = 3时有效',
  `edit_type` int(11) NOT NULL DEFAULT '1' COMMENT '编辑器类型 1 markdown， 2 rich',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '文章状态 1正常，2删除',
  `sort` int(11) NOT NULL COMMENT '排序,值越小越靠前',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


# Dump of table documents
# ------------------------------------------------------------

DROP TABLE IF EXISTS `documents`;

CREATE TABLE `documents` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文档ID',
  `user_id` int(11) NOT NULL COMMENT '创建者',
  `last_user_id` int(11) NOT NULL COMMENT '最后创建者',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '文档标题',
  `domain` varchar(50) NOT NULL DEFAULT '' COMMENT '文档路径名',
  `logo` varchar(100) NOT NULL DEFAULT '' COMMENT '文档LOGO',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '文档描述',
  `permission` int(11) NOT NULL DEFAULT '1' COMMENT '文档权限1公开2内部公开3不公开',
  `password` varchar(64) NOT NULL DEFAULT '' COMMENT '文档密码',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '1 正常  2删除',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_domain` (`domain`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table histories
# ------------------------------------------------------------

DROP TABLE IF EXISTS `histories`;

CREATE TABLE `histories` (
  `hid` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
  `id` int(11) NOT NULL COMMENT '文章ID',
  `pid` int(11) NOT NULL COMMENT '文章父ID',
  `doc_id` int(11) NOT NULL COMMENT '所属文档ID',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '文章标题',
  `user_id` int(11) NOT NULL COMMENT '创建者',
  `last_user_id` int(11) NOT NULL COMMENT '最后编辑者',
  `content` longtext NOT NULL COMMENT '文章内容',
  `type` int(11) NOT NULL COMMENT '文章类型1文章，2外链，3引用，4分类',
  `link` varchar(255) NOT NULL DEFAULT '' COMMENT '链接地址，type=3时有效',
  `reference_id` int(11) NOT NULL COMMENT '引用ID type = 3时有效',
  `edit_type` int(11) NOT NULL COMMENT '编辑器类型 1 markdown， 2 rich',
  `status` int(11) NOT NULL COMMENT '文章状态 1正常，2删除',
  `sort` int(11) NOT NULL COMMENT '排序',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`hid`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table members
# ------------------------------------------------------------

DROP TABLE IF EXISTS `members`;

CREATE TABLE `members` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `doc_id` int(11) NOT NULL COMMENT '文档ID',
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_doc_user` (`doc_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


# Dump of table shares
# ------------------------------------------------------------

DROP TABLE IF EXISTS `shares`;

CREATE TABLE `shares` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '分享名',
  `domain` varchar(50) NOT NULL DEFAULT '' COMMENT '分享文档路径',
  `doc_id` int(11) NOT NULL COMMENT '文档ID',
  `password` varchar(32) NOT NULL DEFAULT '' COMMENT '分享文档密码',
  `share_ids` varchar(255) NOT NULL DEFAULT '' COMMENT '分享文章ID，逗号分隔',
  `user_id` int(11) NOT NULL COMMENT '创建者',
  `status` int(11) NOT NULL COMMENT '1 正常，2删除',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


# Dump of table users
# ------------------------------------------------------------

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `user_name` varchar(50) NOT NULL DEFAULT '' COMMENT '用户名，用于展示',
  `email` varchar(100) NOT NULL DEFAULT '' COMMENT '邮箱',
  `real_name` varchar(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `password` varchar(64) NOT NULL DEFAULT '' COMMENT '密码',
  `avatar` varchar(100) NOT NULL DEFAULT '' COMMENT '用户头像',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '1 启用 2停用',
  `about` varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
  `role` varchar(30) NOT NULL DEFAULT 'user' COMMENT '用户角色admin,leader,user',
  `organization` varchar(50) NOT NULL DEFAULT '' COMMENT '部门组织',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_email` (`email`),
  UNIQUE KEY `un_user_name` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


# Dump of table custom
# ------------------------------------------------------------
DROP TABLE IF EXISTS `custom`;

CREATE TABLE `custom` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(255) NOT NULL COMMENT '帐号',
  `nick_name` varchar(255) NOT NULL COMMENT '昵称',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_name` (`name`),
  UNIQUE KEY `un_nick_name` (`nick_name`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COMMENT='用户表';
