CREATE DATABASE IF NOT EXISTS `xyz_multifinance`;
USE `xyz_multifinance`;

CREATE TABLE IF NOT EXISTS `customers` (
  `id` CHAR(36) PRIMARY KEY,
  `nik` VARCHAR(16) NOT NULL UNIQUE,
  `full_name` VARCHAR(100) NOT NULL,
  `legal_name` VARCHAR(100) NOT NULL,
  `birth_place` VARCHAR(100),
  `birth_date` DATE,
  `salary` DECIMAL(15, 2) NOT NULL,
  `ktp_photo` TEXT,
  `selfie_photo` TEXT,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS `credit_limits` (
  `id` CHAR(36) PRIMARY KEY,
  `customer_id` CHAR(36) NOT NULL,
  `tenor_months` INT NOT NULL,
  `limit_amount` DECIMAL(15, 2) NOT NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE (`customer_id`, `tenor_months`),
  FOREIGN KEY (`customer_id`) REFERENCES `customers` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `transactions` (
  `id` CHAR(36) PRIMARY KEY,
  `customer_id` CHAR(36) NOT NULL,
  `contract_number` VARCHAR(50) UNIQUE NOT NULL,
  `otr_amount` DECIMAL(15, 2) NOT NULL,
  `admin_fee` DECIMAL(15, 2) NOT NULL,
  `installment_amount` DECIMAL(15, 2) NOT NULL,
  `interest_amount` DECIMAL(15, 2) NOT NULL,
  `asset_name` VARCHAR(100),
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (`customer_id`) REFERENCES `customers` (`id`) ON DELETE CASCADE
);
