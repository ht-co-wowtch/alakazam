# Create database
# ------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `platform` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

GRANT ALL ON `platform`.* TO 'default'@'%';
FLUSH PRIVILEGES;
