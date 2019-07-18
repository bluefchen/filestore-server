/*
 Navicat Premium Data Transfer

 Source Server         : connection1
 Source Server Type    : MySQL
 Source Server Version : 80012
 Source Host           : localhost:3306
 Source Schema         : tcpCloudData

 Target Server Type    : MySQL
 Target Server Version : 80012
 File Encoding         : 65001

 Date: 31/03/2019 12:32:58
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for mobileInformation
-- ----------------------------
DROP TABLE IF EXISTS `mobileInformation`;
CREATE TABLE `mobileInformation` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `message` varchar(255) DEFAULT NULL,
  `time` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for sensor
-- ----------------------------
DROP TABLE IF EXISTS `sensor`;
CREATE TABLE `sensor` (
  `sensor_id` int(11) NOT NULL AUTO_INCREMENT,
  `sensor_ip` varchar(20) DEFAULT '0.0.0.0',
  `sensor_name` varchar(32) NOT NULL DEFAULT '传感器',
  `sensor_type` varchar(32) DEFAULT 'xxx传感器',
  `sensor_unit` varchar(10) DEFAULT NULL,
  `isOnline` int(11) DEFAULT '1',
  PRIMARY KEY (`sensor_id`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of sensor
-- ----------------------------
BEGIN;
INSERT INTO `sensor` VALUES (19, '大棚温度', 'T1', '温度传感器', '℃', 1);
INSERT INTO `sensor` VALUES (20, '土壤温度', 'T2', '温度传感器', '℃', 1);
INSERT INTO `sensor` VALUES (21, '大棚照明', 'L', '灯光', 'lx', 1);
INSERT INTO `sensor` VALUES (22, '土壤湿度', 'S', '湿度传感器', '%', 1);
INSERT INTO `sensor` VALUES (23, '二氧化碳浓度', 'C', '二氧化碳传感器', '%', 1);
INSERT INTO `sensor` VALUES (24, '0', '0', '0', '0', 0);
COMMIT;

-- ----------------------------
-- Table structure for sensorInformation
-- ----------------------------
DROP TABLE IF EXISTS `sensorInformation`;
CREATE TABLE `sensorInformation` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `sensor_id` int(11) DEFAULT NULL,
  `message` varchar(255) DEFAULT NULL,
  `time` varchar(32) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `sensor_message_sensor_list_sensor_id_fk` (`sensor_id`),
  CONSTRAINT `sensor_message_sensor_list_sensor_id_fk` FOREIGN KEY (`sensor_id`) REFERENCES `sensor` (`sensor_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2422 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

SET FOREIGN_KEY_CHECKS = 1;
