CREATE DATABASE IF NOT EXISTS `user`;
USE `user`;

CREATE TABLE `user`
(
    `user_id`           int          NOT NULL AUTO_INCREMENT COMMENT '유저 아이디',
    `name`              varchar(10)  NOT NULL COMMENT '이름',     -- 실명 검증 안 된 이름
    `email`             varchar(128) NOT NULL COMMENT '이메일',
    `is_email_verified` tinyint(1)   NOT NULL DEFAULT '0' COMMENT '이메일 검증 여부',
    `phone_number`      varchar(15)  NOT NULL COMMENT '핸드폰 번호', -- format: E.164 (e.g. +821012345678)
    `nickname`          varchar(15)  NOT NULL COMMENT '닉네임',
    `password_hash`     varchar(86)  NOT NULL COMMENT '비밀번호',   -- base64 encoded scrypt output w/ salt: 16 bytes, N: 32768, r: 8, p: 1, len: 32
    `created_at`        timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`),
    UNIQUE KEY `user_u1` (`email`),
    UNIQUE KEY `user_u2` (`phone_number`),
    KEY `user_m1` (`created_at`),
    KEY `user_m2` (`updated_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='유저';


CREATE TABLE `jwt_audience_secret`
(
    `jwt_audience_secret_id` int         NOT NULL AUTO_INCREMENT COMMENT 'JWT audience secret 아이디',
    `audience`               varchar(10) NOT NULL COMMENT 'JWT audience', -- user.user_id
    `secret`                 varchar(44) NOT NULL COMMENT 'secret',       -- base64 encoded 256 bits
    `created_at`             timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`             timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`jwt_audience_secret_id`),
    UNIQUE KEY `jwt_audience_secret_u1` (`audience`),
    KEY `jwt_audience_secret_m1` (`created_at`),
    KEY `jwt_audience_secret_m2` (`updated_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='JWT audience secret';


CREATE TABLE `jwt_denylist`
(
    `jwt_denylist_id` int         NOT NULL AUTO_INCREMENT COMMENT 'JWT denylist 아이디',
    `user_id`         INT         NOT NULL COMMENT '유저 아이디',
    `jti`             VARCHAR(36) NOT NULL COMMENT 'jti(JWT ID) claim', -- uuid v4
    `created_at`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`jwt_denylist_id`),
    UNIQUE KEY `jwt_denylist_u1` (`jti`),
    KEY `jwt_denylist_m1` (`created_at`),
    KEY `jwt_denylist_m2` (`updated_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='JWT denylist';
