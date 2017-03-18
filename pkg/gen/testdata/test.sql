-- MySQL dump 10.13  Distrib 5.7.15, for osx10.12 (x86_64)
--
-- Host: localhost    Database: test
-- ------------------------------------------------------
-- Server version	5.7.17

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `activity_summary`
--

DROP TABLE IF EXISTS `activity_summary`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `activity_summary` (
  `id` char(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `repo_id` int(11) NOT NULL,
  `org_id` int(11) DEFAULT NULL,
  `user_id` int(11) NOT NULL,
  `commits` int(11) DEFAULT NULL,
  `deletions` int(11) DEFAULT NULL,
  `additions` int(11) DEFAULT NULL,
  `issues_created` int(11) DEFAULT NULL,
  `issues_closed` int(11) DEFAULT NULL,
  `issues_commented` int(11) DEFAULT NULL,
  `prs_created` int(11) DEFAULT NULL,
  `prs_closed` int(11) DEFAULT NULL,
  `prs_commented` int(11) DEFAULT NULL,
  `commit_commented` int(11) DEFAULT NULL,
  `day` date NOT NULL,
  `timezone` char(6) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  KEY `activity_summary_day_index` (`day`,`timezone`),
  KEY `activity_summary_repoday_index` (`repo_id`,`day`,`timezone`),
  KEY `activity_summary_user_index` (`user_id`),
  KEY `activity_summary_repo_index` (`repo_id`),
  KEY `activity_summary_userrepo_index` (`user_id`,`repo_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2017-03-17 21:10:59
