CREATE TABLE `users`
(
    `id`         int(11) NOT NULL AUTO_INCREMENT,
    `username`   varchar(255) NOT NULL,
    `password`   varchar(255) NOT NULL,
    `validated`  tinyint(1) NOT NULL,
    `created_at` int(11) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `username` (`username`) USING BTREE
) ENGINE=InnoDB;


CREATE TABLE `banners`
(
    `id`           int(11) NOT NULL AUTO_INCREMENT,
    `img_data`     blob         NOT NULL,
    `banner_text`  varchar(255) NOT NULL,
    `banner_url`   varchar(255) NOT NULL,
    `is_active`    tinyint(1) NOT NULL,
    `limit_shows`  int(11) NOT NULL,
    `limit_clicks` int(11) NOT NULL,
    `limit_budget` int(11) NOT NULL,
    `created_at`   int(11) NOT NULL,
    `user_id`      int(11) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB;