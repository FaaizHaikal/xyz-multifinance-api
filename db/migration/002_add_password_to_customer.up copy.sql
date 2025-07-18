USE `xyz_multifinance`;

ALTER TABLE `Customer`
ADD COLUMN `password` VARCHAR(255) NOT NULL AFTER `full_name`;

ALTER TABLE `Customer` ADD INDEX `idx_nik_password` (`nik`, `password`);
