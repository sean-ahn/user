CREATE DATABASE IF NOT EXISTS `user`;
USE `user`;

CREATE TABLE `user`
(
    `user_id`           int          NOT NULL AUTO_INCREMENT COMMENT '유저 아이디',
    `name`              varchar(10)  NOT NULL COMMENT '이름',     -- NOTE: may not be real name
    `email`             varchar(128) NOT NULL COMMENT '이메일',
    `is_email_verified` tinyint(1)   NOT NULL COMMENT '이메일 검증 여부',
    `phone_number`      varchar(15)  NOT NULL COMMENT '핸드폰 번호', -- format: E.164 (e.g. +821012345678)
    `nickname`          varchar(15)  NOT NULL COMMENT '닉네임',
    `password_hash`     varchar(86)  NOT NULL COMMENT '비밀번호',   -- format: base64 encoded 32 bytes scrypt output w/ salt: 16 bytes, N: 32768, r: 8, p: 1, keyLen: 32
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
    `audience`               varchar(10) NOT NULL COMMENT 'JWT audience claim',
    `secret`                 varchar(44) NOT NULL COMMENT 'secret', -- format: base64 encoded 256 bits
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
    `user_id`         int         NOT NULL COMMENT '유저 아이디',            -- user.user_id
    `jti`             varchar(36) NOT NULL COMMENT 'jti(JWT ID) claim', -- format: uuid v4
    `created_at`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`jwt_denylist_id`),
    UNIQUE KEY `jwt_denylist_u1` (`jti`),
    KEY `jwt_denylist_m1` (`created_at`),
    KEY `jwt_denylist_m2` (`updated_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='JWT denylist';


CREATE TABLE `sms_otp_verification`
(
    `sms_otp_verification_id`  int AUTO_INCREMENT COMMENT 'SMS OTP 인증 아이디',
    `verification_token`       varchar(36) NOT NULL COMMENT '인증 토큰',  -- format: uuid v4
    `phone_number`             varchar(15) NOT NULL COMMENT '핸드폰 번호', -- format: E.164 (e.g. +821012345678)
    `otp_code`                 varchar(6)  NOT NULL COMMENT '인증 코드',
    `expires_at`               timestamp   NOT NULL COMMENT '만료 일시',
    `verification_trials`      int(11)     NOT NULL COMMENT '검증 시도 횟수',
    `verification_valid_until` timestamp            DEFAULT NULL COMMENT '검증 유효 일시',
    `created_at`               timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`               timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`sms_otp_verification_id`),
    UNIQUE KEY `sms_otp_verification_u1` (`verification_token`),
    KEY `sms_otp_verification_m1` (`created_at`),
    KEY `sms_otp_verification_m2` (`updated_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='SMS OTP 인증';
