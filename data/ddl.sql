CREATE DATABASE IF NOT EXISTS `user`;
USE `user`;

CREATE TABLE `user`
(
    `user_id`            int          NOT NULL AUTO_INCREMENT COMMENT '유저 아이디',
    `name`               varchar(10)  NOT NULL COMMENT '이름',     -- 실명 검증 안 된 이름
    `email`              varchar(128) NOT NULL COMMENT '이메일',
    `is_email_confirmed` tinyint(1)   NOT NULL DEFAULT '0' COMMENT '이메일 검증 여부',
    `phone_number`       varchar(15)  NOT NULL COMMENT '핸드폰 번호', -- format: E.164 (e.g. 821012345678)
    `nickname`           varchar(15)  NOT NULL COMMENT '닉네임',
    `password_hashed`    varchar(82)  NOT NULL COMMENT '비밀번호',   -- max length scrypt 256 bits
    `created_at`         timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`),
    UNIQUE KEY `user_u1` (`email`),
    UNIQUE KEY `user_u2` (`phone_number`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='유저';
