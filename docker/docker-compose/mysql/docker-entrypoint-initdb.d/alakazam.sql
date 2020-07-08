# Create database
# ------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `alakazam` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

GRANT ALL ON `alakazam`.* TO 'default'@'%';
FLUSH PRIVILEGES;
