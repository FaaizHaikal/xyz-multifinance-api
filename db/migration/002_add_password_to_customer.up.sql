USE `xyz_multifinance`;

ALTER TABLE `customers`
ADD COLUMN `password` VARCHAR(255) NOT NULL AFTER `full_name`;

ALTER TABLE `customers` ADD INDEX `idx_nik_password` (`nik`, `password`);
