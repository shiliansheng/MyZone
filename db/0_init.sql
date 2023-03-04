-- CREATE USER 'bee'@'%' IDENTIFIED BY 'bee';
-- CREATE DATABASE localdata DEFAULT CHARSET utf8 COLLATE utf8_general_ci;
-- GRANT ALL ON localdata.* TO 'bee'@'%';
CREATE DATABASE `myzone` default charset utf8 COLLATE utf8_general_ci;

USE `myzone`;

DROP TABLE IF EXISTS `category`;
CREATE TABLE `category`(
    id      int AUTO_INCREMENT  COMMENT '编号',
    title   varchar(256)    NOT NULL,
    `state` int             NOT NULL DEFAULT '0' COMMENT'0:valid,1:invalid',
    `path`  varchar(256)    NOT NULL,
    `module` int NOT NULL DEFAULT '1',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="类型表";

DROP TABLE IF EXISTS `album`;
CREATE TABLE `album`(
    id      int AUTO_INCREMENT COMMENT '编号',
    title   varchar(256) NOT NULL,
    `path`  varchar(256) NOT NULL,
    categoryid  int NOT NULL,
    collect int NOT NULL DEFAULT '0',
    `state` int NOT NULL DEFAULT '0' COMMENT'0:valid,1:invalid',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="图集表";

DROP TABLE IF EXISTS `picture`;
CREATE TABLE `picture`(
    id      int AUTO_INCREMENT COMMENT '编号',
    title   varchar(256),
    albumid int NOT NULL,
    collect int NOT NULL DEFAULT '0',
    `path`  varchar(256) NOT NULL,
    `state` int NOT NULL DEFAULT '0' COMMENT'0:valid,1:invalid',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="图片表";

DROP TABLE IF EXISTS `video`;
CREATE TABLE `video`(
    id      int AUTO_INCREMENT COMMENT '编号',
    title   varchar(256),
    cover    varchar(256),
    `path`  varchar(256) NOT NULL,
    categoryid  int NOT NULL DEFAULT '0',
    tagid   varchar(128) DEFAULT '',
    collect int NOT NULL DEFAULT '0',
    view    int NOT NULL DEFAULT '0',
    serialid int  DEFAULT '0',
    actorid varchar(128) DEFAULT '',
    duration varchar(16) DEFAULT '00:00:00',
    pubtime datetime DEFAULT CURRENT_TIMESTAMP,
    `state` int NOT NULL DEFAULT '0' COMMENT'0:valid,1:invalid',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="视频表";

-- alter table video add column timenode varchar(256) default '[]';

DROP TABLE IF EXISTS `serial`;
CREATE TABLE `serial`(
    id      int AUTO_INCREMENT COMMENT '编号',
    title   varchar(256),
    collect int NOT NULL DEFAULT '0',
    `count` int NOT NULL DEFAULT '0',
    categoryid  int NOT NULL DEFAULT '0',
    cover    varchar(256) NOT NULL,
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="系列";

DROP TABLE IF EXISTS `actor`;
CREATE TABLE `actor`(
    id      int AUTO_INCREMENT COMMENT '编号',
    `name`  varchar(128) NOT NULL UNIQUE,
    collect int NOT NULL DEFAULT '0',
    cover    varchar(256) NOT NULL,
    pubtime timestamp DEFAULT CURRENT_TIMESTAMP,
    recommend int DEFAULT '0',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="演员表";

DROP TABLE IF EXISTS `tag`;
CREATE TABLE `tag`(
    id      int AUTO_INCREMENT COMMENT '编号',
    `name`   varchar(128) NOT NULL UNIQUE,
    recommend int NOT NULL DEFAULT '0',
    -- categoryid  int NOT NULL DEFAULT '0',
    -- viewnum int NOT NULL DEFAULT '0',
    -- addnum  int NOT NULL DEFAULT '0',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="标签表";


DROP TABLE IF EXISTS `spider`;
CREATE TABLE `spider`(
    `id`       int AUTO_INCREMENT NOT NULL,
	`title`    varchar(256) NOT NULL DEFAULT '',
	`url`      varchar(256) NOT NULL DEFAULT '',
	`pubtime`  varchar(32),
    `datastr`  varchar(512) DEFAULT '',
	`module`   int DEFAULT '0',
	`section`  int DEFAULT '0',
	`category` int DEFAULT '0',
	`view`     int DEFAULT '0',
	`addtime`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `preid`    int NOT NULL DEFAULT '0',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="爬虫表";

DROP TABLE IF EXISTS `record`;
CREATE TABLE `record`(
    `id`       int AUTO_INCREMENT NOT NULL,
	`title`    varchar(256) NOT NULL DEFAULT '',
	`detail`    varchar(256) DEFAULT '',
	`category` int DEFAULT '0',
    `top`       int default '0',
    `content` varchar(2048) DEFAULT '',
	`addtime`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`deltime`  varchar(32) DEFAULT '',
    `state` int NOT NULL DEFAULT '0' COMMENT'0:valid,1:invalid',
    PRIMARY KEY(id)
)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT="记录表";